package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/FriedGlue/BookIt/api/pkg/shared"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/dynamodb/expression"
)

func SearchBooks(request events.APIGatewayProxyRequest) events.APIGatewayProxyResponse {
	// Grab query params, e.g. ?isbn=XXX or ?q=someTitle
	isbn := request.QueryStringParameters["isbn"]
	q := request.QueryStringParameters["q"]
	bookId := request.QueryStringParameters["bookId"]
	openLibraryId := request.QueryStringParameters["openLibraryId"]

	// If you want separate params for title, author, etc. you could parse them here:
	// titleParam := request.QueryStringParameters["title"]
	// authorParam := request.QueryStringParameters["author"]
	// etc.

	if isbn == "" && q == "" && bookId == "" && openLibraryId == "" {
		return shared.ErrorResponse(400, "Please provide at least one search parameter (?isbn= or ?q= or ?bookId= or ?openLibraryId=).")
	}

	svc := shared.DynamoDBClient()

	var books []BookData
	var err error

	if bookId != "" {
		books, err = searchByBookId(svc, bookId)
		if err != nil {
			log.Printf("Error searching by bookId: %v\n", err)
			return shared.ErrorResponse(500, fmt.Sprintf("Error searching by bookId: %v", err))
		}
	} else if isbn != "" {
		books, err = searchByISBN(svc, isbn)
		if err != nil {
			log.Printf("Error searching by ISBN: %v\n", err)
			return shared.ErrorResponse(500, fmt.Sprintf("Error searching by ISBN: %v", err))
		}
	} else if openLibraryId != "" {
		books, err = searchByOpenLibraryId(svc, openLibraryId)
		if err != nil {
			log.Printf("Error searching by openLibraryId: %v\n", err)
			return shared.ErrorResponse(500, fmt.Sprintf("Error searching by openLibraryId: %v", err))
		}
	} else {
		// 2) If a general query (q) is provided, do a partial match on 'titleLowercase' or do a scan:
		if q != "" {
			// For small scale, do a scan.
			// Example: 'contains(titleLowercase, :qLower)'
			books, err = searchByPartialTitle(svc, q)
			if err != nil {
				log.Printf("Error searching by partial title: %v\n", err)
				return shared.ErrorResponse(500, fmt.Sprintf("Error searching by partial title: %v", err))
			}
		}
	}

	// Convert to JSON
	responseBytes, _ := json.Marshal(books)

	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Body:       string(responseBytes),
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
	}
}

// Exact ISBN lookup using GSI
func searchByISBN(svc *dynamodb.DynamoDB, isbnValue string) ([]BookData, error) {
	input := &dynamodb.QueryInput{
		TableName:              aws.String(BOOKS_TABLE_NAME),
		IndexName:              aws.String(ISBN_INDEX_NAME), // "ISBNIndex"
		KeyConditionExpression: aws.String("isbn13 = :isbnVal"),
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":isbnVal": {S: aws.String(isbnValue)},
		},
	}

	result, err := svc.Query(input)
	if err != nil {
		return nil, err
	}

	var books []BookData
	if err := dynamodbattribute.UnmarshalListOfMaps(result.Items, &books); err != nil {
		return nil, err
	}

	return books, nil
}

// Exact bookId lookup using primary key
func searchByBookId(svc *dynamodb.DynamoDB, bookId string) ([]BookData, error) {
	input := &dynamodb.GetItemInput{
		TableName: aws.String(BOOKS_TABLE_NAME),
		Key: map[string]*dynamodb.AttributeValue{
			"bookId": {S: aws.String(bookId)},
		},
	}

	result, err := svc.GetItem(input)
	if err != nil {
		return nil, err
	}

	var book BookData
	if err := dynamodbattribute.UnmarshalMap(result.Item, &book); err != nil {
		return nil, err
	}

	return []BookData{book}, nil
}

// Exact bookId lookup using primary key
func searchByOpenLibraryId(svc *dynamodb.DynamoDB, openLibraryId string) ([]BookData, error) {
	log.Printf("Searching for book with OpenLibrary ID: %s", openLibraryId)

	input := &dynamodb.QueryInput{
		TableName:              aws.String(BOOKS_TABLE_NAME),
		IndexName:              aws.String(OPEN_LIBRARY_INDEX_NAME), // "OpenLibraryIndex"
		KeyConditionExpression: aws.String("openLibraryId = :openLibraryId"),
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":openLibraryId": {S: aws.String(openLibraryId)},
		},
	}

	result, err := svc.Query(input)
	if err != nil {
		return nil, err
	}

	var books []BookData
	if err := dynamodbattribute.UnmarshalListOfMaps(result.Items, &books); err != nil {
		return nil, err
	}

	if len(books) == 0 {
		return nil, fmt.Errorf("no book found with OpenLibrary ID: %s", openLibraryId)
	}

	// Return just the first book since we expect only one match
	return []BookData{books[0]}, nil
}

// Partial match on 'titleLowercase' with "contains(...)"
func searchByPartialTitle(svc *dynamodb.DynamoDB, query string) ([]BookData, error) {
	lowerQuery := strings.ToLower(query)
	// Build filter expression: "contains(titleLowercase, :q)"
	filt := expression.Contains(expression.Name("titleLowercase"), lowerQuery)

	expr, err := expression.NewBuilder().WithFilter(filt).Build()
	if err != nil {
		return nil, err
	}

	input := &dynamodb.ScanInput{
		TableName:                 aws.String(BOOKS_TABLE_NAME),
		FilterExpression:          expr.Filter(),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
	}

	result, err := svc.Scan(input)
	if err != nil {
		return nil, err
	}

	var books []BookData
	if err := dynamodbattribute.UnmarshalListOfMaps(result.Items, &books); err != nil {
		return nil, err
	}

	return books, nil
}
