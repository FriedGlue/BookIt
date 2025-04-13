package handlers

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/FriedGlue/BookIt/api/pkg/shared"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/google/uuid"
)

// OpenLibraryResponse represents the data structure returned from Open Library API
type OpenLibraryResponse struct {
	Description   Description    `json:"description"`
	Title         string         `json:"title"`
	URL           string         `json:"url"`
	NumberOfPages int            `json:"number_of_pages"`
	Subjects      []string       `json:"subjects"`
	PublishDate   string         `json:"publish_date"`
	Authors       []Author       `json:"authors"`
	PublishPlaces []PublishPlace `json:"publish_places"`
	Cover         []Cover        `json:"cover"`
	Publishers    []Publisher    `json:"publishers"`
	Identifiers   Identifiers    `json:"identifiers"`
}

type Description struct {
	Value string `json:"value"`
	Type  string `json:"type"`
}

type Publisher struct {
	Name string `json:"name"`
}

type Author struct {
	URL  string `json:"url"`
	Name string `json:"name"`
}

type Excerpt struct {
	Comment string `json:"comment"`
	Text    string `json:"text"`
}

type PublishPlace struct {
	Name string `json:"name"`
}

type Cover struct {
	Small  string `json:"small"`
	Medium string `json:"medium"`
	Large  string `json:"large"`
}

type Identifiers struct {
	ISBN13       []string `json:"isbn_13"`
	ISBN10       []string `json:"isbn_10"`
	Google       []string `json:"google"`
	LCCN         []string `json:"lccn"`
	OCLC         []string `json:"oclc"`
	Goodreads    []string `json:"goodreads"`
	LibraryThing []string `json:"librarything"`
}

// SaveExternalBook handles POST requests to /books/save-external-book
func SaveExternalBook(request events.APIGatewayProxyRequest) events.APIGatewayProxyResponse {
	// First try to extract user info using the standard method
	userId, err := shared.GetUserIDFromToken(request)
	if err != nil {
		log.Printf("Authentication error: %v", err)
		return shared.ErrorResponse(401, "Authentication required")
	}

	// Log user making the request for debugging
	log.Printf("Processing save-external-book request from user: %s", userId)

	// Parse request body
	var requestBody struct {
		BookId string `json:"bookId"`
	}

	err = json.Unmarshal([]byte(request.Body), &requestBody)
	if err != nil {
		return shared.ErrorResponse(400, "Invalid request body")
	}

	// Log the request for debugging
	log.Printf("Processing save-external-book request for bookId: %s", requestBody.BookId)

	// Determine if this is an Open Library ID
	bookId := requestBody.BookId
	if len(bookId) > 2 && bookId[:2] == "OL" {
		// This is an Open Library ID
		// Check if the book already exists in our database
		svc := shared.DynamoDBClient()
		existingBooks, err := searchByOpenLibraryId(svc, bookId)
		if err == nil && len(existingBooks) > 0 {
			// Book already exists, return it
			log.Printf("Book already exists in database: %s", bookId)
			return shared.SuccessResponse(200, existingBooks[0])
		}

		// Book doesn't exist, fetch from Open Library
		log.Printf("Fetching book from Open Library: %s", bookId)
		bookData, err := fetchBookFromOpenLibrary(bookId)
		if err != nil {
			log.Printf("Error fetching from Open Library: %v", err)
			return shared.ErrorResponse(500, fmt.Sprintf("Error fetching book from Open Library: %v", err))
		}

		// Save the book to our database
		log.Printf("Saving book to database: %s", bookId)
		savedBook, err := saveBookToDynamoDB(svc, bookData, bookId, userId)
		if err != nil {
			log.Printf("Error saving to database: %v", err)
			return shared.ErrorResponse(500, fmt.Sprintf("Error saving book to database: %v", err))
		}

		// Return the saved book
		log.Printf("Successfully saved book: %s", bookId)
		return shared.SuccessResponse(200, savedBook)
	} else {
		// This is not an Open Library ID, try to find it in our database
		svc := shared.DynamoDBClient()
		books, err := searchByBookId(svc, bookId)
		if err != nil || len(books) == 0 {
			log.Printf("Book not found in database: %s", bookId)
			return shared.ErrorResponse(404, "Book not found")
		}

		// Return the found book
		log.Printf("Found book in database: %s", bookId)
		return shared.SuccessResponse(200, books[0])
	}
}

// fetchBookFromOpenLibrary fetches book data from Open Library API
func fetchBookFromOpenLibrary(openLibraryId string) (*OpenLibraryResponse, error) {
	// Format: OL12345W -> /works/OL12345W
	url := fmt.Sprintf("https://openlibrary.org/works/%s.json?jscmd=data", openLibraryId)

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch book: %s", resp.Status)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// The API returns a JSON object where the key is the ISBN or ID
	var responseMap map[string]*OpenLibraryResponse
	err = json.Unmarshal(body, &responseMap)
	if err != nil {
		// Try parsing as a direct object in case the API response format changes
		var bookData OpenLibraryResponse
		err = json.Unmarshal(body, &bookData)
		if err != nil {
			return nil, err
		}
		return &bookData, nil
	}

	// Extract the first book from the map
	for _, bookData := range responseMap {
		return bookData, nil
	}

	return nil, fmt.Errorf("no book data found in the response")
}

// saveBookToDynamoDB saves the book data to DynamoDB
func saveBookToDynamoDB(svc *dynamodb.DynamoDB, bookData *OpenLibraryResponse, openLibraryId string, userId string) (*BookData, error) {
	// Prepare book data for DynamoDB
	newBook := BookData{
		BookID:         uuid.New().String(),
		Title:          bookData.Title,
		TitleLowercase: strings.ToLower(bookData.Title),
		PageCount:      bookData.NumberOfPages,
		OpenLibraryId:  openLibraryId,
		Tags:           []string{},
		Description:    bookData.Description.Value,
	}

	if bookData.Description.Value != "" {
		newBook.Description = bookData.Description.Value
	}

	// Set a default page count if it's zero
	if newBook.PageCount == 0 {
		// Default to 300 pages if we don't have the actual count
		newBook.PageCount = 300
		log.Printf("Setting default page count (300) for book with no page information: %s", openLibraryId)
	}

	// Add publish date as a tag if available
	if bookData.PublishDate != "" {
		newBook.Tags = append(newBook.Tags, "Published:"+bookData.PublishDate)
	}

	// Set cover image URL if available
	if len(bookData.Cover) > 0 && bookData.Cover[0].Large != "" {
		newBook.CoverImageURL = bookData.Cover[0].Large
	} else if len(bookData.Cover) > 0 && bookData.Cover[0].Medium != "" {
		newBook.CoverImageURL = bookData.Cover[0].Medium
	} else if len(bookData.Cover) > 0 && bookData.Cover[0].Small != "" {
		newBook.CoverImageURL = bookData.Cover[0].Small
	}

	// Set ISBNs if available
	if len(bookData.Identifiers.ISBN13) > 0 {
		newBook.ISBN13 = bookData.Identifiers.ISBN13[0]
	}
	if len(bookData.Identifiers.ISBN10) > 0 {
		newBook.ISBN10 = bookData.Identifiers.ISBN10[0]
	}

	// Get author names if available
	if len(bookData.Authors) > 0 {
		var authorNames []string
		for _, author := range bookData.Authors {
			if author.Name != "" {
				authorNames = append(authorNames, author.Name)
			}
		}
		if len(authorNames) > 0 {
			newBook.Authors = authorNames
		}
	}

	// Add subjects as tags
	if len(bookData.Subjects) > 0 {
		for _, subject := range bookData.Subjects {
			if subject != "" {
				newBook.Tags = append(newBook.Tags, subject)
			}
		}
	}

	// Add publisher information as a tag
	if len(bookData.Publishers) > 0 && bookData.Publishers[0].Name != "" {
		newBook.Tags = append(newBook.Tags, "Publisher:"+bookData.Publishers[0].Name)
	}

	// Create additional identifiers as tags
	addIdentifierAsTag := func(idType string, values []string) {
		if len(values) > 0 {
			newBook.Tags = append(newBook.Tags, fmt.Sprintf("%s:%s", idType, values[0]))
		}
	}

	addIdentifierAsTag("Google", bookData.Identifiers.Google)
	addIdentifierAsTag("LCCN", bookData.Identifiers.LCCN)
	addIdentifierAsTag("OCLC", bookData.Identifiers.OCLC)
	addIdentifierAsTag("Goodreads", bookData.Identifiers.Goodreads)
	addIdentifierAsTag("LibraryThing", bookData.Identifiers.LibraryThing)

	// Always add the OpenLibrary tag
	newBook.Tags = append(newBook.Tags, "OpenLibrary:"+openLibraryId)

	// Convert book to DynamoDB attribute value
	item, err := dynamodbattribute.MarshalMap(newBook)
	if err != nil {
		return nil, err
	}

	// Save to DynamoDB
	input := &dynamodb.PutItemInput{
		TableName: aws.String(BOOKS_TABLE_NAME),
		Item:      item,
	}

	_, err = svc.PutItem(input)
	if err != nil {
		return nil, err
	}

	return &newBook, nil
}
