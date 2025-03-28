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
	Title         string   `json:"title"`
	Authors       []Author `json:"authors"`
	Covers        []int    `json:"covers"`
	NumberOfPages int      `json:"number_of_pages"`
	ISBN13        []string `json:"isbn_13"`
	Description   string   `json:"description,omitempty"`
}

type Author struct {
	Key string `json:"key"`
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
	url := fmt.Sprintf("https://openlibrary.org/works/%s.json", openLibraryId)

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

	var bookData OpenLibraryResponse
	err = json.Unmarshal(body, &bookData)
	if err != nil {
		return nil, err
	}

	return &bookData, nil
}

// saveBookToDynamoDB saves the book data to DynamoDB
func saveBookToDynamoDB(svc *dynamodb.DynamoDB, bookData *OpenLibraryResponse, openLibraryId string, userId string) (*BookData, error) {
	// Prepare book data for DynamoDB
	newBook := BookData{
		BookID:         uuid.New().String(),
		Title:          bookData.Title,
		TitleLowercase: strings.ToLower(bookData.Title),
		PageCount:      bookData.NumberOfPages,
		// e OpenLibraryID in our schema, but we can store it as a tag
		Tags: []string{"OpenLibrary:" + openLibraryId},
	}

	// Set a default page count if it's zero
	if newBook.PageCount == 0 {
		// Default to 300 pages if we don't have the actual count
		newBook.PageCount = 300
		log.Printf("Setting default page count (300) for book with no page information: %s", openLibraryId)
	}

	// Set cover image URL if available
	if len(bookData.Covers) > 0 {
		newBook.CoverImageURL = fmt.Sprintf("https://covers.openlibrary.org/b/id/%d-L.jpg", bookData.Covers[0])
	}

	// Set ISBN if available
	if len(bookData.ISBN13) > 0 {
		newBook.ISBN13 = bookData.ISBN13[0]
	}

	// Set description if available
	if bookData.Description != "" {
		if descriptionField, ok := interface{}(newBook).(interface{ SetDescription(string) }); ok {
			descriptionField.SetDescription(bookData.Description)
		}
	}

	// Get author names if available
	if len(bookData.Authors) > 0 {
		authorNames, err := fetchAuthorNames(bookData.Authors)
		if err == nil && len(authorNames) > 0 {
			newBook.Authors = authorNames
		}
	}

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

// fetchAuthorNames fetches author names from Open Library API
func fetchAuthorNames(authors []Author) ([]string, error) {
	var authorNames []string

	for _, author := range authors {
		// Format: /authors/OL12345A -> OL12345A
		authorId := author.Key
		if len(authorId) > 9 {
			authorId = authorId[9:]
		}

		url := fmt.Sprintf("https://openlibrary.org/authors/%s.json", authorId)

		resp, err := http.Get(url)
		if err != nil {
			log.Printf("Error fetching author %s: %v", authorId, err)
			continue
		}

		if resp.StatusCode != http.StatusOK {
			log.Printf("Failed to fetch author %s: %s", authorId, resp.Status)
			resp.Body.Close()
			continue
		}

		body, err := ioutil.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			log.Printf("Error reading author response for %s: %v", authorId, err)
			continue
		}

		var authorData struct {
			Name string `json:"name"`
		}

		err = json.Unmarshal(body, &authorData)
		if err != nil {
			log.Printf("Error parsing author data for %s: %v", authorId, err)
			continue
		}

		if authorData.Name != "" {
			authorNames = append(authorNames, authorData.Name)
		}
	}

	return authorNames, nil
}
