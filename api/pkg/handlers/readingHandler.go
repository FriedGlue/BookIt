package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/dynamodb/expression"
)

// For this example, we'll call our DynamoDB table "Books"
// with a partition key named "isbn" (string).

// Book represents a single book record in DynamoDB.
type Book struct {
	ISBN            string   `json:"isbn"` // Partition key
	Title           string   `json:"title,omitempty"`
	Authors         []string `json:"authors,omitempty"`
	PageCount       int      `json:"pageCount,omitempty"`
	PublicationDate string   `json:"publicationDate,omitempty"`
	CoverImageURL   string   `json:"coverImageUrl,omitempty"`
	Tags            []string `json:"tags,omitempty"`
	// Add other fields as desired (genres, mood, pace, etc.)
}

// DynamoDBClient initializes a DynamoDB client session.
func DynamoDBClient() *dynamodb.DynamoDB {
	sess := session.Must(session.NewSession())
	return dynamodb.New(sess)
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
func fetchBookFromOpenLibrary(isbn string) (Book, error) {
	// Example: https://openlibrary.org/api/books?bibkeys=ISBN:<isbn>&format=json&jscmd=data
	url := fmt.Sprintf("https://openlibrary.org/api/books?bibkeys=ISBN:%s&format=json&jscmd=data", isbn)

	resp, err := http.Get(url)
	if err != nil {
		return Book{}, fmt.Errorf("HTTP request error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return Book{}, fmt.Errorf("Open Library returned status code %d", resp.StatusCode)
	}

	var responseMap map[string]OLBookData
	if err := json.NewDecoder(resp.Body).Decode(&responseMap); err != nil {
		return Book{}, fmt.Errorf("JSON decode error: %v", err)
	}

	key := fmt.Sprintf("ISBN:%s", isbn)
	olData, exists := responseMap[key]
	if !exists {
		return Book{}, fmt.Errorf("No data found for ISBN %s", isbn)
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

	b := Book{
		ISBN:            isbn,
		Title:           olData.Title,
		Authors:         authors,
		PageCount:       olData.NumberOfPages,
		PublicationDate: olData.PublishDate,
		CoverImageURL:   coverURL,
		Tags:            subjects,
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
	isbn, hasISBN := request.PathParameters["isbn"]
	svc := DynamoDBClient()

	if hasISBN && isbn != "" {
		// Retrieve a single book by primary key
		getInput := &dynamodb.GetItemInput{
			TableName: aws.String("Books"),
			Key: map[string]*dynamodb.AttributeValue{
				"isbn": {S: aws.String(isbn)},
			},
		}
		result, err := svc.GetItem(getInput)
		if err != nil {
			return errorResponse(500, "Error retrieving book: "+err.Error())
		}
		if result.Item == nil {
			return errorResponse(404, "Book not found")
		}

		var book Book
		err = dynamodbattribute.UnmarshalMap(result.Item, &book)
		if err != nil {
			return errorResponse(500, "Error unmarshalling book: "+err.Error())
		}

		bytes, _ := json.Marshal(book)
		return events.APIGatewayProxyResponse{
			StatusCode: 200,
			Body:       string(bytes),
		}
	}

	// No ISBN => we might do a full table scan or return an error
	// WARNING: Doing a scan on a large table can be expensive
	scanInput := &dynamodb.ScanInput{
		TableName: aws.String("Books"),
	}
	scanRes, err := svc.Scan(scanInput)
	if err != nil {
		return errorResponse(500, "Scan error: "+err.Error())
	}
	if len(scanRes.Items) == 0 {
		return errorResponse(404, "No books found")
	}

	var books []Book
	err = dynamodbattribute.UnmarshalListOfMaps(scanRes.Items, &books)
	if err != nil {
		return errorResponse(500, "Error unmarshalling items: "+err.Error())
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
		return errorResponse(400, "Invalid JSON request: "+err.Error())
	}
	if strings.TrimSpace(input.ISBN) == "" {
		return errorResponse(400, "ISBN is required in POST body")
	}

	// 1) Fetch data from Open Library
	book, err := fetchBookFromOpenLibrary(input.ISBN)
	if err != nil {
		return errorResponse(500, "Failed to fetch from Open Library: "+err.Error())
	}

	// 2) Store in DynamoDB
	av, err := dynamodbattribute.MarshalMap(book)
	if err != nil {
		return errorResponse(500, "Error marshalling book data: "+err.Error())
	}

	svc := DynamoDBClient()
	putInput := &dynamodb.PutItemInput{
		TableName: aws.String("Books"),
		Item:      av,
	}
	_, err = svc.PutItem(putInput)
	if err != nil {
		return errorResponse(500, "DynamoDB PutItem error: "+err.Error())
	}

	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Body:       fmt.Sprintf("Book with ISBN %s created successfully", book.ISBN),
	}
}

// 3. PUT /books/{isbn}
//   - The request body can contain partial updates.
//   - We’ll demonstrate using a DynamoDB UpdateItem with an expression builder.
func UpdateBook(request events.APIGatewayProxyRequest) events.APIGatewayProxyResponse {
	isbn, hasISBN := request.PathParameters["isbn"]
	if !hasISBN || isbn == "" {
		return errorResponse(400, "Missing path parameter: isbn")
	}

	// This struct can contain any fields you want to allow updating
	var updates struct {
		Title           *string   `json:"title,omitempty"`
		Authors         *[]string `json:"authors,omitempty"`
		PageCount       *int      `json:"pageCount,omitempty"`
		PublicationDate *string   `json:"publicationDate,omitempty"`
		CoverImageURL   *string   `json:"coverImageUrl,omitempty"`
		Tags            *[]string `json:"tags,omitempty"`
	}

	if err := json.Unmarshal([]byte(request.Body), &updates); err != nil {
		return errorResponse(400, "Invalid JSON request: "+err.Error())
	}

	// Convert partial updates to map
	updateMap, err := dynamodbattribute.MarshalMap(updates)
	if err != nil {
		return errorResponse(500, "Error marshalling update data: "+err.Error())
	}

	// Build update expression
	updateBuilder := expression.UpdateBuilder{}
	for key, val := range updateMap {
		// If val is not null, set in expression
		if val.NULL == nil {
			updateBuilder = updateBuilder.Set(expression.Name(key), expression.Value(val))
		}
	}

	expr, err := expression.NewBuilder().WithUpdate(updateBuilder).Build()
	if err != nil {
		return errorResponse(500, "Error building expression: "+err.Error())
	}

	svc := DynamoDBClient()
	input := &dynamodb.UpdateItemInput{
		TableName: aws.String("Books"),
		Key: map[string]*dynamodb.AttributeValue{
			"isbn": {S: aws.String(isbn)},
		},
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		UpdateExpression:          expr.Update(),
		ReturnValues:              aws.String("UPDATED_NEW"),
	}

	_, err = svc.UpdateItem(input)
	if err != nil {
		return errorResponse(500, "Error updating book: "+err.Error())
	}

	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Body:       fmt.Sprintf("Book with ISBN %s updated successfully", isbn),
	}
}

// 4. DELETE /books/{isbn}
func DeleteBook(request events.APIGatewayProxyRequest) events.APIGatewayProxyResponse {
	isbn, hasISBN := request.PathParameters["isbn"]
	if !hasISBN || isbn == "" {
		return errorResponse(400, "Missing path parameter: isbn")
	}

	svc := DynamoDBClient()
	input := &dynamodb.DeleteItemInput{
		TableName: aws.String("Books"),
		Key: map[string]*dynamodb.AttributeValue{
			"isbn": {S: aws.String(isbn)},
		},
	}

	_, err := svc.DeleteItem(input)
	if err != nil {
		return errorResponse(500, "Error deleting book: "+err.Error())
	}

	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Body:       fmt.Sprintf("Book with ISBN %s deleted successfully", isbn),
	}
}

// ===============================
//  Utility Functions
// ===============================

// errorResponse is a helper to generate an APIGatewayProxyResponse with a given status and message.
func errorResponse(status int, message string) events.APIGatewayProxyResponse {
	return events.APIGatewayProxyResponse{
		StatusCode: status,
		Body:       message,
	}
}
