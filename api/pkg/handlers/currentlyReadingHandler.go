package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"math/rand"
	"time"

	"github.com/FriedGlue/BookIt/api/pkg/models"
	"github.com/FriedGlue/BookIt/api/pkg/shared"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

// ----------------------- Handlers -----------------------

// GetCurrentlyReading retrieves the "currently reading" list from the Profile table
func GetCurrentlyReading(request events.APIGatewayProxyRequest) events.APIGatewayProxyResponse {
	log.Println("GetCurrentlyReading invoked")
	userId, err := shared.GetUserIDFromToken(request)
	if err != nil {
		log.Printf("Error extracting userId: %v\n", err)
		return errorResponse(401, err.Error())
	}

	svc := DynamoDBClient()
	input := &dynamodb.GetItemInput{
		TableName: aws.String(PROFILES_TABLE_NAME),
		Key: map[string]*dynamodb.AttributeValue{
			"_id": {S: aws.String(userId)},
		},
	}

	log.Printf("Fetching profile for userId: %s\n", userId)
	result, err := svc.GetItem(input)
	if err != nil {
		log.Printf("DynamoDB GetItem error: %v\n", err)
		return errorResponse(500, fmt.Sprintf("DynamoDB GetItem error: %v", err))
	}
	if result.Item == nil {
		log.Println("Profile not found")
		return errorResponse(404, "Profile not found")
	}

	var profile models.Profile
	if err := dynamodbattribute.UnmarshalMap(result.Item, &profile); err != nil {
		log.Printf("Error unmarshalling profile: %v\n", err)
		return errorResponse(500, "Error unmarshalling profile: "+err.Error())
	}

	responseBody, err := json.Marshal(profile.CurrentlyReading)
	if err != nil {
		log.Printf("Error marshalling response: %v\n", err)
		return errorResponse(500, "Error marshalling currently reading response")
	}

	log.Println("Currently reading list retrieval successful")
	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Body:       string(responseBody),
	}
}

type newCurrentlyReadingItemRequest struct {
	ISBN   string `json:"isbn"`
	BookID string `json:"bookId,omitempty"`
}

// AddToCurrentlyReading adds a new currentlyReadingItem to the "currently reading" list in the Profile table
func AddToCurrentlyReading(request events.APIGatewayProxyRequest) events.APIGatewayProxyResponse {
	log.Println("AddToCurrentlyReading invoked")
	userId, err := shared.GetUserIDFromToken(request)
	if err != nil {
		log.Printf("Error extracting userId: %v\n", err)
		return errorResponse(401, err.Error())
	}

	var newCurrentlyReadingItemRequest newCurrentlyReadingItemRequest
	if err := json.Unmarshal([]byte(request.Body), &newCurrentlyReadingItemRequest); err != nil {
		log.Printf("Invalid JSON: %v\n", err)
		return errorResponse(400, "Invalid JSON: "+err.Error())
	}

	svc := DynamoDBClient()
	getInput := &dynamodb.GetItemInput{
		TableName: aws.String(PROFILES_TABLE_NAME),
		Key: map[string]*dynamodb.AttributeValue{
			"_id": {S: aws.String(userId)},
		},
	}

	result, err := svc.GetItem(getInput)
	if err != nil {
		log.Printf("DynamoDB GetItem error: %v\n", err)
		return errorResponse(500, fmt.Sprintf("DynamoDB GetItem error: %v", err))
	}
	if result.Item == nil {
		log.Println("Profile not found")
		return errorResponse(404, "Profile not found")
	}

	var profile models.Profile
	if err := dynamodbattribute.UnmarshalMap(result.Item, &profile); err != nil {
		log.Printf("Error unmarshalling profile: %v\n", err)
		return errorResponse(500, "Error unmarshalling profile: "+err.Error())
	}

	found := false
	for _, item := range profile.CurrentlyReading {
		if item.Book.ISBN == newCurrentlyReadingItemRequest.ISBN || item.Book.BookID == newCurrentlyReadingItemRequest.BookID {
			found = true
			break
		}
	}

	if found {
		return errorResponse(409, "Book already in currently reading list")
	}

	var bookDetails BookData
	if newCurrentlyReadingItemRequest.BookID != "" {
		// If the book ID is provided, we need to fetch the book details from the Books table
		getBookInput := &dynamodb.GetItemInput{
			TableName: aws.String(BOOKS_TABLE_NAME),
			Key: map[string]*dynamodb.AttributeValue{
				"bookId": {S: aws.String(newCurrentlyReadingItemRequest.BookID)},
			},
		}
		bookResult, err := svc.GetItem(getBookInput)
		if err != nil {
			log.Printf("DynamoDB GetItem error: %v\n", err)
			return errorResponse(500, fmt.Sprintf("DynamoDB GetItem error: %v", err))
		}
		if bookResult.Item == nil {
			log.Println("Book not found")
			return errorResponse(404, "Book not found")
		}
		if err := dynamodbattribute.UnmarshalMap(bookResult.Item, &bookDetails); err != nil {
			log.Printf("Error unmarshalling book details: %v\n", err)
			return errorResponse(500, "Error unmarshalling book details: "+err.Error())
		}
	} else {
		// If the book ID is not provided, we need to fetch the book details from the openlibrary API
		bookDetails, err = FetchBookFromOpenLibrary(newCurrentlyReadingItemRequest.ISBN)
		if err != nil {
			log.Printf("Error fetching book details: %v\n", err)
			return errorResponse(500, "Error fetching book details: "+err.Error())
		}
	}

	// Create a new CurrentlyReadingItem and add it to the profile using the book details
	temp := rand.New(rand.NewSource(time.Now().UnixNano()))
	book := models.Book{
		BookID:     fmt.Sprintf("%d", temp.Int()),
		ISBN:       bookDetails.ISBN,
		Title:      bookDetails.Title,
		Authors:    bookDetails.Authors,
		CoverImage: bookDetails.CoverImageURL,
		TotalPages: bookDetails.PageCount,
		Progress: models.ReadingProgress{
			LastPageRead: 0,
			Percentage:   0,
			LastUpdated:  time.Now().Format(time.RFC3339),
		},
	}
	currentlyReadingItem := models.CurrentlyReadingItem{
		Book:        book,
		StartedDate: time.Now().Format(time.RFC3339),
	}

	profile.CurrentlyReading = append(profile.CurrentlyReading, currentlyReadingItem)

	updatedProfile, err := dynamodbattribute.MarshalMap(profile)
	if err != nil {
		log.Printf("Error marshalling updated profile: %v\n", err)
		return errorResponse(500, "Error marshalling updated profile: "+err.Error())
	}

	putInput := &dynamodb.PutItemInput{
		TableName: aws.String(PROFILES_TABLE_NAME),
		Item:      updatedProfile,
	}
	_, err = svc.PutItem(putInput)
	if err != nil {
		log.Printf("DynamoDB PutItem error: %v\n", err)
		return errorResponse(500, fmt.Sprintf("DynamoDB PutItem error: %v", err))
	}

	log.Printf("Book added to currently reading for user %s\n", userId)
	return events.APIGatewayProxyResponse{
		StatusCode: 201,
		Body:       "Book added to currently reading",
	}
}

// UpdateCurrentlyReading updates a book in the "currently reading" list in the Profile table
type updateCurrentlyReadingRequest struct {
	ISBN        string `json:"isbn"`
	CurrentPage int    `json:"currentPage"`
	BookID      string `json:"bookId,omitempty"`
	Title       string `json:"title,omitempty"`
}

// UpdateCurrentlyReading updates a book in the "currently reading" list in the Profile table
func UpdateCurrentlyReading(request events.APIGatewayProxyRequest) events.APIGatewayProxyResponse {
	log.Println("UpdateCurrentlyReading invoked")
	userId, err := shared.GetUserIDFromToken(request)
	if err != nil {
		log.Printf("Error extracting userId: %v\n", err)
		return errorResponse(401, err.Error())
	}

	var updateCurrentlyReadingRequest updateCurrentlyReadingRequest
	if err := json.Unmarshal([]byte(request.Body), &updateCurrentlyReadingRequest); err != nil {
		log.Printf("Invalid JSON: %v\n", err)
		return errorResponse(400, "Invalid JSON: "+err.Error())
	}

	svc := DynamoDBClient()
	getInput := &dynamodb.GetItemInput{
		TableName: aws.String(PROFILES_TABLE_NAME),
		Key: map[string]*dynamodb.AttributeValue{
			"_id": {S: aws.String(userId)},
		},
	}

	result, err := svc.GetItem(getInput)
	if err != nil {
		log.Printf("DynamoDB GetItem error: %v\n", err)
		return errorResponse(500, fmt.Sprintf("DynamoDB GetItem error: %v", err))
	}
	if result.Item == nil {
		log.Println("Profile not found")
		return errorResponse(404, "Profile not found")
	}

	var profile models.Profile
	if err := dynamodbattribute.UnmarshalMap(result.Item, &profile); err != nil {
		log.Printf("Error unmarshalling profile: %v\n", err)
		return errorResponse(500, "Error unmarshalling profile: "+err.Error())
	}

	found := false
	var storedBook models.CurrentlyReadingItem

	// need to add check to ensure only one book is allowed inside the currently reading list
	for i, item := range profile.CurrentlyReading {
		if item.Book.ISBN == updateCurrentlyReadingRequest.ISBN {
			storedBook = profile.CurrentlyReading[i]
			found = true
			break
		}
	}

	if !found {
		return errorResponse(404, "Book not found in currently reading list")
	}

	newProgressPercentage := math.Floor(float64(float64(updateCurrentlyReadingRequest.CurrentPage) / float64(storedBook.Book.TotalPages) * 100))

	storedBook.Book.Progress.LastPageRead = updateCurrentlyReadingRequest.CurrentPage
	storedBook.Book.Progress.Percentage = newProgressPercentage
	storedBook.Book.Progress.LastUpdated = time.Now().Format(time.RFC3339)

	for i, item := range profile.CurrentlyReading {
		if item.Book.ISBN == updateCurrentlyReadingRequest.ISBN {
			profile.CurrentlyReading[i].Book.Progress = storedBook.Book.Progress
			break
		}
	}

	updatedProfile, err := dynamodbattribute.MarshalMap(profile)
	if err != nil {
		log.Printf("Error marshalling updated profile: %v\n", err)
		return errorResponse(500, "Error marshalling updated profile: "+err.Error())
	}

	putInput := &dynamodb.PutItemInput{
		TableName: aws.String(PROFILES_TABLE_NAME),
		Item:      updatedProfile,
	}
	_, err = svc.PutItem(putInput)
	if err != nil {
		log.Printf("DynamoDB PutItem error: %v\n", err)
		return errorResponse(500, fmt.Sprintf("DynamoDB PutItem error: %v", err))
	}

	log.Printf("Book updated in currently reading for user %s\n", userId)
	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Body:       "Book updated in currently reading",
	}
}

// RemoveFromCurrentlyReading removes a book from the "currently reading" list in the Profile table
func RemoveFromCurrentlyReading(request events.APIGatewayProxyRequest) events.APIGatewayProxyResponse {
	log.Println("RemoveFromCurrentlyReading invoked")
	userId, err := shared.GetUserIDFromToken(request)
	if err != nil {
		log.Printf("Error extracting userId: %v\n", err)
		return errorResponse(401, err.Error())
	}

	bookID := request.QueryStringParameters["bookID"]
	isbn := request.QueryStringParameters["isbn"]
	if bookID == "" && isbn == "" {
		return errorResponse(400, "bookID or isbn query parameter are required")
	}

	svc := DynamoDBClient()
	getInput := &dynamodb.GetItemInput{
		TableName: aws.String(PROFILES_TABLE_NAME),
		Key: map[string]*dynamodb.AttributeValue{
			"_id": {S: aws.String(userId)},
		},
	}

	result, err := svc.GetItem(getInput)
	if err != nil {
		log.Printf("DynamoDB GetItem error: %v\n", err)
		return errorResponse(500, fmt.Sprintf("DynamoDB GetItem error: %v", err))
	}
	if result.Item == nil {
		log.Println("Profile not found")
		return errorResponse(404, "Profile not found")
	}

	var profile models.Profile
	if err := dynamodbattribute.UnmarshalMap(result.Item, &profile); err != nil {
		log.Printf("Error unmarshalling profile: %v\n", err)
		return errorResponse(500, "Error unmarshalling profile: "+err.Error())
	}

	index := -1
	for i, item := range profile.CurrentlyReading {
		if item.Book.BookID == bookID || item.Book.ISBN == bookID {
			index = i
			break
		}
	}

	if index == -1 {
		return errorResponse(404, "Book not found in currently reading list")
	}

	profile.CurrentlyReading = append(profile.CurrentlyReading[:index], profile.CurrentlyReading[index+1:]...)

	updatedProfile, err := dynamodbattribute.MarshalMap(profile)
	if err != nil {
		log.Printf("Error marshalling updated profile: %v\n", err)
		return errorResponse(500, "Error marshalling updated profile: "+err.Error())
	}

	putInput := &dynamodb.PutItemInput{
		TableName: aws.String(PROFILES_TABLE_NAME),
		Item:      updatedProfile,
	}
	_, err = svc.PutItem(putInput)
	if err != nil {
		log.Printf("DynamoDB PutItem error: %v\n", err)
		return errorResponse(500, fmt.Sprintf("DynamoDB PutItem error: %v", err))
	}

	log.Printf("Book removed from currently reading for user %s\n", userId)
	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Body:       "Book removed from currently reading",
	}
}
