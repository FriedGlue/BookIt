package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/FriedGlue/BookIt/api/pkg/shared"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/dynamodb/expression"
	"github.com/google/uuid"
)

// For this example, we'll call our DynamoDB table "Books"
// with a partition key named "isbn" (string).

var (
	PROFILES_TABLE_NAME     = os.Getenv("PROFILES_TABLE_NAME")     // e.g. "UserProfiles"
	BOOKS_TABLE_NAME        = os.Getenv("BOOKS_TABLE_NAME")        // e.g. "UserProfiles"
	OPEN_LIBRARY_INDEX_NAME = os.Getenv("OPEN_LIBRARY_INDEX_NAME") // e.g. "OpenLibraryIndex"
	ISBN_INDEX_NAME         = os.Getenv("ISBN_INDEX_NAME")         // e.g. "ISBNIndex"
)

// Book represents a single book record in DynamoDB.
type BookData struct {
	BookID         string   `json:"bookId"`
	ISBN10         string   `json:"isbn10,omitempty"`
	ISBN13         string   `json:"isbn13,omitempty"`
	Title          string   `json:"title,omitempty"`
	TitleLowercase string   `json:"titleLowercase,omitempty"`
	Authors        []string `json:"authors,omitempty"`
	PageCount      int      `json:"pageCount,omitempty"`
	CoverImageURL  string   `json:"coverImageUrl,omitempty"`
	Tags           []string `json:"tags,omitempty"`
}

// ===============================
//  Open Library Fetch Logic
// ===============================

// OLBookData shapes the "api/books?bibkeys=..." response from Open Library.
type OLBookData struct {
	Title         string      `json:"title"`
	Authors       []OLAuthor  `json:"authors"`
	NumberOfPages int         `json:"number_of_pages"`
	PublishDate   string      `json:"publish_date"`
	Cover         OLCover     `json:"cover"`
	Subjects      []OLSubject `json:"subjects"`
}

type OLAuthor struct {
	Name string `json:"name"`
}

type OLCover struct {
	Small  string `json:"small"`
	Medium string `json:"medium"`
	Large  string `json:"large"`
}

type OLSubject struct {
	Name string `json:"name"`
}

// fetchBookFromOpenLibrary fetches metadata by ISBN via Open Library.
func FetchBookFromOpenLibrary(isbn string) (BookData, error) {
	// Example: https://openlibrary.org/api/books?bibkeys=ISBN:<isbn>&format=json&jscmd=data
	url := fmt.Sprintf("https://openlibrary.org/api/books?bibkeys=ISBN:%s&format=json&jscmd=data", isbn)

	resp, err := http.Get(url)
	if err != nil {
		return BookData{}, fmt.Errorf("HTTP request error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return BookData{}, fmt.Errorf("Open Library returned status code %d", resp.StatusCode)
	}

	var responseMap map[string]OLBookData
	if err := json.NewDecoder(resp.Body).Decode(&responseMap); err != nil {
		return BookData{}, fmt.Errorf("JSON decode error: %v", err)
	}

	key := fmt.Sprintf("ISBN:%s", isbn)
	olData, exists := responseMap[key]
	if !exists {
		return BookData{}, fmt.Errorf("No data found for ISBN %s", isbn)
	}

	// Convert from OLBookData -> our Book struct
	var authors []string
	for _, a := range olData.Authors {
		authors = append(authors, a.Name)
	}

	// Attempt to pick the largest cover image
	coverURL := olData.Cover.Large
	if coverURL == "" {
		coverURL = olData.Cover.Medium
	}
	if coverURL == "" {
		coverURL = olData.Cover.Small
	}

	var subjects []string
	for _, s := range olData.Subjects {
		subjects = append(subjects, s.Name)
	}

	// Generate a unique BookID using UUID
	bookId := uuid.New().String()

	b := BookData{
		BookID:         bookId,
		ISBN13:         isbn,
		Title:          olData.Title,
		TitleLowercase: strings.ToLower(olData.Title),
		Authors:        authors,
		PageCount:      olData.NumberOfPages,
		CoverImageURL:  coverURL,
		Tags:           subjects,
	}
	return b, nil
}

// ===============================
//  CRUD Handlers (API Gateway)
// ===============================

// 1. GET /books/{isbn?}
//   - If an ISBN is provided in pathParameters["isbn"], retrieve that specific book.
//   - If no ISBN is provided, do a scan or handle accordingly (not recommended at large scale).
func GetBooks(request events.APIGatewayProxyRequest) events.APIGatewayProxyResponse {
	bookId, hasBookId := request.PathParameters["isbn"]
	svc := shared.DynamoDBClient()

	if hasBookId && bookId != "" {
		// Retrieve a single book by primary key
		getInput := &dynamodb.GetItemInput{
			TableName: aws.String(BOOKS_TABLE_NAME),
			Key: map[string]*dynamodb.AttributeValue{
				"bookId": {S: aws.String(bookId)},
			},
		}
		result, err := svc.GetItem(getInput)
		if err != nil {
			return shared.ErrorResponse(500, "Error retrieving book: "+err.Error())
		}
		if result.Item == nil {
			return shared.ErrorResponse(404, "Book not found")
		}

		var book BookData
		err = dynamodbattribute.UnmarshalMap(result.Item, &book)
		if err != nil {
			return shared.ErrorResponse(500, "Error unmarshalling book: "+err.Error())
		}

		bytes, _ := json.Marshal(book)
		return events.APIGatewayProxyResponse{
			StatusCode: 200,
			Body:       string(bytes),
		}
	}

	// No bookId => we might do a full table scan or return an error
	// WARNING: Doing a scan on a large table can be expensive
	scanInput := &dynamodb.ScanInput{
		TableName: aws.String(BOOKS_TABLE_NAME),
	}
	scanRes, err := svc.Scan(scanInput)
	if err != nil {
		return shared.ErrorResponse(500, "Scan error: "+err.Error())
	}
	if len(scanRes.Items) == 0 {
		return shared.ErrorResponse(404, "No books found")
	}

	var books []BookData
	err = dynamodbattribute.UnmarshalListOfMaps(scanRes.Items, &books)
	if err != nil {
		return shared.ErrorResponse(500, "Error unmarshalling items: "+err.Error())
	}

	bytes, _ := json.Marshal(books)
	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Body:       string(bytes),
	}
}

// 2. POST /books
//   - The request body can include a JSON with an `isbn` field to fetch from Open Library,
//     or a full Book object to store directly.
func CreateBook(request events.APIGatewayProxyRequest) events.APIGatewayProxyResponse {
	var input struct {
		ISBN string `json:"isbn"`
		// Optionally, you could allow manual override of some fields
	}

	if err := json.Unmarshal([]byte(request.Body), &input); err != nil {
		return shared.ErrorResponse(400, "Invalid JSON request: "+err.Error())
	}
	if strings.TrimSpace(input.ISBN) == "" {
		return shared.ErrorResponse(400, "ISBN is required in POST body")
	}

	// 1) Fetch data from Open Library
	book, err := FetchBookFromOpenLibrary(input.ISBN)
	if err != nil {
		return shared.ErrorResponse(500, "Failed to fetch from Open Library: "+err.Error())
	}

	// 2) Store in DynamoDB
	av, err := dynamodbattribute.MarshalMap(book)
	if err != nil {
		return shared.ErrorResponse(500, "Error marshalling book data: "+err.Error())
	}

	svc := shared.DynamoDBClient()
	putInput := &dynamodb.PutItemInput{
		TableName: aws.String(BOOKS_TABLE_NAME),
		Item:      av,
	}
	_, err = svc.PutItem(putInput)
	if err != nil {
		return shared.ErrorResponse(500, "DynamoDB PutItem error: "+err.Error())
	}

	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Body:       fmt.Sprintf("Book with ISBN %s created successfully", book.ISBN13),
	}
}

// 3. PUT /books/{bookId}
//   - The request body can contain partial updates.
//   - We'll demonstrate using a DynamoDB UpdateItem with an expression builder.
func UpdateBook(request events.APIGatewayProxyRequest) events.APIGatewayProxyResponse {
	bookId, hasBookId := request.PathParameters["bookId"]
	if !hasBookId || bookId == "" {
		return shared.ErrorResponse(400, "Missing path parameter: bookId")
	}

	// This struct matches the updatable fields in BookData
	var updates struct {
		ISBN10        *string   `json:"isbn10,omitempty"`
		ISBN13        *string   `json:"isbn13,omitempty"`
		Title         *string   `json:"title,omitempty"`
		Authors       *[]string `json:"authors,omitempty"`
		PageCount     *int      `json:"pageCount,omitempty"`
		CoverImageURL *string   `json:"coverImageUrl,omitempty"`
		Tags          *[]string `json:"tags,omitempty"`
	}

	if err := json.Unmarshal([]byte(request.Body), &updates); err != nil {
		return shared.ErrorResponse(400, "Invalid JSON request: "+err.Error())
	}

	// Convert partial updates to map
	updateMap, err := dynamodbattribute.MarshalMap(updates)
	if err != nil {
		return shared.ErrorResponse(500, "Error marshalling update data: "+err.Error())
	}

	// Build update expression
	updateBuilder := expression.UpdateBuilder{}
	for key, val := range updateMap {
		if val.NULL == nil {
			// Special handling for title to also update titleLowercase
			if key == "title" {
				updateBuilder = updateBuilder.Set(expression.Name(key), expression.Value(val))
				// Add titleLowercase field update
				if val.S != nil {
					updateBuilder = updateBuilder.Set(
						expression.Name("titleLowercase"),
						expression.Value(strings.ToLower(*val.S)),
					)
				}
			} else {
				updateBuilder = updateBuilder.Set(expression.Name(key), expression.Value(val))
			}
		}
	}

	expr, err := expression.NewBuilder().WithUpdate(updateBuilder).Build()
	if err != nil {
		return shared.ErrorResponse(500, "Error building expression: "+err.Error())
	}

	svc := shared.DynamoDBClient()
	input := &dynamodb.UpdateItemInput{
		TableName: aws.String(BOOKS_TABLE_NAME),
		Key: map[string]*dynamodb.AttributeValue{
			"bookId": {S: aws.String(bookId)},
		},
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		UpdateExpression:          expr.Update(),
		ReturnValues:              aws.String("UPDATED_NEW"),
	}

	_, err = svc.UpdateItem(input)
	if err != nil {
		return shared.ErrorResponse(500, "Error updating book: "+err.Error())
	}

	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Body:       fmt.Sprintf("Book with ID %s updated successfully", bookId),
	}
}

// 4. DELETE /books/{isbn}
func DeleteBook(request events.APIGatewayProxyRequest) events.APIGatewayProxyResponse {
	isbn, hasISBN := request.QueryStringParameters["isbn"]
	if !hasISBN || isbn == "" {
		return shared.ErrorResponse(400, "Missing query string parameter: isbn")
	}

	svc := shared.DynamoDBClient()
	input := &dynamodb.DeleteItemInput{
		TableName: aws.String(BOOKS_TABLE_NAME),
		Key: map[string]*dynamodb.AttributeValue{
			"isbn": {S: aws.String(isbn)},
		},
	}

	_, err := svc.DeleteItem(input)
	if err != nil {
		return shared.ErrorResponse(500, "Error deleting book: "+err.Error())
	}

	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Body:       fmt.Sprintf("Book with ISBN %s deleted successfully", isbn),
	}
}
