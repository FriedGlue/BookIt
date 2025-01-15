package handlers

import (
	"encoding/json"
	"fmt"
	"log"

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

// AddToCurrentlyReading adds a book to the "currently reading" list in the Profile table
func AddToCurrentlyReading(request events.APIGatewayProxyRequest) events.APIGatewayProxyResponse {
	log.Println("AddToCurrentlyReading invoked")
	userId, err := shared.GetUserIDFromToken(request)
	if err != nil {
		log.Printf("Error extracting userId: %v\n", err)
		return errorResponse(401, err.Error())
	}

	var book models.CurrentlyReadingItem
	if err := json.Unmarshal([]byte(request.Body), &book); err != nil {
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

	profile.CurrentlyReading = append(profile.CurrentlyReading, book)

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
func UpdateCurrentlyReading(request events.APIGatewayProxyRequest) events.APIGatewayProxyResponse {
	log.Println("UpdateCurrentlyReading invoked")
	userId, err := shared.GetUserIDFromToken(request)
	if err != nil {
		log.Printf("Error extracting userId: %v\n", err)
		return errorResponse(401, err.Error())
	}

	var book models.CurrentlyReadingItem
	if err := json.Unmarshal([]byte(request.Body), &book); err != nil {
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

	updated := false
	for i, item := range profile.CurrentlyReading {
		if item.Book.BookID == book.Book.BookID {
			profile.CurrentlyReading[i] = book
			updated = true
			break
		}
	}

	if !updated {
		return errorResponse(404, "Book not found in currently reading list")
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
	if bookID == "" {
		return errorResponse(400, "bookID query parameter is required")
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
		if item.Book.BookID == bookID {
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
