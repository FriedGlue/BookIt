package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"regexp"
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

	// Check if openLibraryId is actually a UUID format
	if openLibraryId != "" {
		// Basic UUID format validation regex
		// Format: 8-4-4-4-12 hexadecimal characters
		isUuidFormat := false
		if len(openLibraryId) == 36 {
			// Check if it matches UUID format
			matched, _ := regexp.MatchString(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`, strings.ToLower(openLibraryId))
			isUuidFormat = matched
		}

		if isUuidFormat {
			log.Printf("Detected UUID format in openLibraryId parameter: %s, redirecting to bookId search", openLibraryId)
			bookId = openLibraryId
			openLibraryId = ""
		}
	}

	if bookId != "" {
		books, err = searchByBookId(svc, bookId)
		if err != nil {
			log.Printf("Error searching by bookId: %v\n", err)
			return shared.ErrorResponse(500, fmt.Sprintf("Error searching by bookId: %v", err))
		}
		if len(books) == 0 {
			return shared.ErrorResponse(404, fmt.Sprintf("No book found with ID: %s", bookId))
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

	// Check if the item was found
	if result.Item == nil || len(result.Item) == 0 {
		return []BookData{}, nil // Return empty slice instead of error
	}

	var book BookData
	if err := dynamodbattribute.UnmarshalMap(result.Item, &book); err != nil {
		return nil, err
	}

	return []BookData{book}, nil
}

// Search for books by OpenLibraryId
func searchByOpenLibraryId(svc *dynamodb.DynamoDB, openLibraryId string) ([]BookData, error) {
	log.Printf("Searching for book with OpenLibrary ID: %s", openLibraryId)

	// Try two approaches:
	// 1. First, look for the OpenLibraryId in dedicated field
	// 2. As fallback, check for tag with OpenLibrary prefix

	// Build filter expression for scan: "openLibraryId = :olid OR contains(tags, :tagPrefix)"
	olTagPrefix := "OpenLibrary:" + openLibraryId

	// First try with the dedicated field
	filt := expression.Equal(expression.Name("openLibraryId"), expression.Value(openLibraryId))

	expr, err := expression.NewBuilder().WithFilter(filt).Build()
	if err != nil {
		return nil, err
	}

	scanInput := &dynamodb.ScanInput{
		TableName:                 aws.String(BOOKS_TABLE_NAME),
		FilterExpression:          expr.Filter(),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
	}

	result, err := svc.Scan(scanInput)
	if err != nil {
		return nil, err
	}

	var books []BookData
	if err := dynamodbattribute.UnmarshalListOfMaps(result.Items, &books); err != nil {
		return nil, err
	}

	// If we found books, return them
	if len(books) > 0 {
		log.Printf("Found book with OpenLibrary ID %s using openLibraryId field", openLibraryId)
		return []BookData{books[0]}, nil
	}

	// Otherwise, try the tag-based approach as fallback
	log.Printf("No book found with openLibraryId field, trying tag-based search for: %s", olTagPrefix)

	// Build filter to search in tags array
	tagFilt := expression.Contains(expression.Name("tags"), olTagPrefix)

	tagExpr, err := expression.NewBuilder().WithFilter(tagFilt).Build()
	if err != nil {
		return nil, err
	}

	tagScanInput := &dynamodb.ScanInput{
		TableName:                 aws.String(BOOKS_TABLE_NAME),
		FilterExpression:          tagExpr.Filter(),
		ExpressionAttributeNames:  tagExpr.Names(),
		ExpressionAttributeValues: tagExpr.Values(),
	}

	tagResult, err := svc.Scan(tagScanInput)
	if err != nil {
		return nil, err
	}

	var tagBooks []BookData
	if err := dynamodbattribute.UnmarshalListOfMaps(tagResult.Items, &tagBooks); err != nil {
		return nil, err
	}

	if len(tagBooks) == 0 {
		return nil, fmt.Errorf("no book found with OpenLibrary ID: %s", openLibraryId)
	}

	log.Printf("Found book with OpenLibrary ID %s using tags search", openLibraryId)
	return []BookData{tagBooks[0]}, nil
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
