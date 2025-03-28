package handlers

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/FriedGlue/BookIt/api/pkg/shared"
	"github.com/aws/aws-lambda-go/events"
)

// OpenLibrarySearchResponse represents the search results from Open Library API
type OpenLibrarySearchResponse struct {
	NumFound int                      `json:"num_found"`
	Start    int                      `json:"start"`
	Docs     []OpenLibrarySearchEntry `json:"docs"`
}

// OpenLibrarySearchEntry represents a single result in the search response
type OpenLibrarySearchEntry struct {
	Key          string   `json:"key"`
	Title        string   `json:"title"`
	AuthorName   []string `json:"author_name,omitempty"`
	CoverI       int      `json:"cover_i,omitempty"`
	ISBN         []string `json:"isbn,omitempty"`
	FirstPublish int      `json:"first_publish_year,omitempty"`
}

// SearchResultEntry represents a combined search result entry
type SearchResultEntry struct {
	BookId        string   `json:"bookId"`
	Title         string   `json:"title"`
	Authors       []string `json:"authors,omitempty"`
	Thumbnail     string   `json:"thumbnail,omitempty"`
	Source        string   `json:"source"`
	OpenLibraryId string   `json:"openLibraryId,omitempty"`
}

// SearchOpenLibrary searches the Open Library API
func SearchOpenLibrary(query string) ([]SearchResultEntry, error) {
	if len(query) < 2 {
		// Skip very short queries to prevent unnecessary API calls
		return []SearchResultEntry{}, nil
	}

	// URL encode the query
	encodedQuery := url.QueryEscape(query)
	url := fmt.Sprintf("https://openlibrary.org/search.json?q=%s&limit=10", encodedQuery)

	// Add timeout to prevent hanging requests
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	resp, err := client.Get(url)
	if err != nil {
		log.Printf("Error fetching from Open Library: %v", err)
		return []SearchResultEntry{}, nil // Return empty results instead of error
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("Open Library API returned status code: %d", resp.StatusCode)
		return []SearchResultEntry{}, nil // Return empty results on error
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error reading Open Library response: %v", err)
		return []SearchResultEntry{}, nil
	}

	var searchResponse OpenLibrarySearchResponse
	if err := json.Unmarshal(body, &searchResponse); err != nil {
		log.Printf("Error parsing Open Library response: %v", err)
		return []SearchResultEntry{}, nil
	}

	var results []SearchResultEntry
	for _, doc := range searchResponse.Docs {
		// Extract the Open Library ID from the key
		// Format: /works/OL12345W -> OL12345W
		olId := ""
		if len(doc.Key) > 7 {
			olId = doc.Key[7:] // Remove "/works/"
		}

		// Skip entries without a valid ID
		if olId == "" {
			continue
		}

		// Build the thumbnail URL if a cover ID is available
		thumbnail := ""
		if doc.CoverI > 0 {
			thumbnail = fmt.Sprintf("https://covers.openlibrary.org/b/id/%d-M.jpg", doc.CoverI)
		}

		result := SearchResultEntry{
			BookId:        olId, // Using Open Library ID as the book ID for external books
			Title:         doc.Title,
			Authors:       doc.AuthorName,
			Thumbnail:     thumbnail,
			Source:        "openLibrary",
			OpenLibraryId: olId,
		}
		results = append(results, result)
	}

	log.Printf("Open Library search for '%s' returned %d results", query, len(results))
	return results, nil
}

// MergeSearchResults combines results from our database and Open Library
func MergeSearchResults(dbResults []BookData, olResults []SearchResultEntry) []SearchResultEntry {
	var mergedResults []SearchResultEntry

	// First, convert our database results to the common format
	for _, book := range dbResults {
		thumbnail := book.CoverImageURL
		result := SearchResultEntry{
			BookId:        book.BookID,
			Title:         book.Title,
			Authors:       book.Authors,
			Thumbnail:     thumbnail,
			Source:        "database",
			OpenLibraryId: book.OpenLibraryId,
		}
		mergedResults = append(mergedResults, result)
	}

	// Track Open Library IDs that are already in our database
	olIdsInDb := make(map[string]bool)
	for _, book := range dbResults {
		if book.OpenLibraryId != "" {
			olIdsInDb[book.OpenLibraryId] = true
		}
	}

	// Add Open Library results that aren't already in our database
	for _, olResult := range olResults {
		if _, exists := olIdsInDb[olResult.OpenLibraryId]; !exists {
			mergedResults = append(mergedResults, olResult)
		}
	}

	return mergedResults
}

// CombinedSearch searches both our database and Open Library
func CombinedSearch(request events.APIGatewayProxyRequest) events.APIGatewayProxyResponse {
	// Extract the query parameter
	q := request.QueryStringParameters["q"]
	if q == "" {
		return shared.ErrorResponse(400, "Please provide a search query using the 'q' parameter")
	}

	log.Printf("Combined search request for query: '%s'", q)

	// Extract user info
	userId, err := shared.GetUserIDFromToken(request)
	if err != nil {
		log.Printf("Warning: Unauthenticated search request: %v", err)
		// Continue with the search even without authentication
	} else {
		log.Printf("Search request from user: %s", userId)
	}

	// Search our database
	svc := shared.DynamoDBClient()
	dbBooks, err := searchByPartialTitle(svc, q)
	if err != nil {
		log.Printf("Error searching database: %v", err)
		// Continue with Open Library search even if database search fails
		dbBooks = []BookData{}
	}

	log.Printf("Database search returned %d results", len(dbBooks))

	// Search Open Library if query is not too short
	var olResults []SearchResultEntry
	if len(q) >= 2 {
		olResults, err = SearchOpenLibrary(q)
		if err != nil {
			log.Printf("Error searching Open Library: %v", err)
			// Continue with database results even if Open Library search fails
			olResults = []SearchResultEntry{}
		}
		log.Printf("Open Library search returned %d results", len(olResults))
	} else {
		log.Printf("Query too short for Open Library search, skipping")
	}

	// Merge the results
	mergedResults := MergeSearchResults(dbBooks, olResults)

	log.Printf("Merged search results: %d total items", len(mergedResults))

	// Convert to JSON and return
	responseBytes, _ := json.Marshal(mergedResults)
	return shared.SuccessResponse(200, json.RawMessage(responseBytes))
}
