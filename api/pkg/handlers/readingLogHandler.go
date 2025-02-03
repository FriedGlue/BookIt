package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/FriedGlue/BookIt/api/pkg/models"
	"github.com/FriedGlue/BookIt/api/pkg/shared"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

// ProfilesTableName is read from the environment.
var ProfilesTableName = os.Getenv("PROFILES_TABLE_NAME")

// HandleGetReadingLog is a Lambda handler to retrieve a user's reading log.
func GetReadingLog(request events.APIGatewayProxyRequest) events.APIGatewayProxyResponse {
	log.Println("GetReadingLog invoked")
	userId, err := shared.GetUserIDFromToken(request)
	if err != nil {
		log.Printf("Error extracting userId: %v\n", err)
		return shared.ErrorResponse(401, err.Error())
	}

	svc := shared.DynamoDBClient()
	getInput := &dynamodb.GetItemInput{
		TableName: aws.String(ProfilesTableName),
		Key: map[string]*dynamodb.AttributeValue{
			"_id": {S: aws.String(userId)},
		},
	}

	result, err := svc.GetItem(getInput)
	if err != nil {
		return shared.ErrorResponse(500, fmt.Sprintf("Error retrieving profile: %v", err))
	}
	if result.Item == nil {
		return shared.ErrorResponse(404, "Profile not found")
	}

	var profile models.Profile
	if err := dynamodbattribute.UnmarshalMap(result.Item, &profile); err != nil {
		return shared.ErrorResponse(500, fmt.Sprintf("Error unmarshalling profile: %v", err))
	}

	// Marshal the reading log to JSON.
	responseBody, err := json.Marshal(profile.ReadingLog)
	if err != nil {
		return shared.ErrorResponse(500, fmt.Sprintf("Error marshalling reading log: %v", err))
	}

	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Body:       string(responseBody),
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
	}
}

type UpdateReadingLogItemRequest struct {
	ReadingLogItemId string `json:"readingLogItemId"`
	PagesRead        int    `json:"pagesRead,omitempty"`
	Notes            string `json:"notes,omitempty"`
}

func UpdateReadingLogItem(request events.APIGatewayProxyRequest) events.APIGatewayProxyResponse {
	log.Println("UpdateReadingLog invoked")

	// Extract the user ID from the token.
	userId, err := shared.GetUserIDFromToken(request)
	if err != nil {
		log.Printf("Error extracting userId: %v\n", err)
		return shared.ErrorResponse(401, err.Error())
	}

	// Retrieve the user's profile from DynamoDB.
	svc := shared.DynamoDBClient()
	getInput := &dynamodb.GetItemInput{
		TableName: aws.String("PROFILES_TABLE_NAME"),
		Key: map[string]*dynamodb.AttributeValue{
			"_id": {S: aws.String(userId)},
		},
	}

	result, err := svc.GetItem(getInput)
	if err != nil {
		return shared.ErrorResponse(500, fmt.Sprintf("Error retrieving profile: %v", err))
	}
	if result.Item == nil {
		return shared.ErrorResponse(404, "Profile not found")
	}

	var profile models.Profile
	if err := dynamodbattribute.UnmarshalMap(result.Item, &profile); err != nil {
		return shared.ErrorResponse(500, fmt.Sprintf("Error unmarshalling profile: %v", err))
	}

	// Unmarshal the request body into our update request structure.
	var updateReq UpdateReadingLogItemRequest
	if err := json.Unmarshal([]byte(request.Body), &updateReq); err != nil {
		return shared.ErrorResponse(400, fmt.Sprintf("Invalid request body: %v", err))
	}

	if updateReq.ReadingLogItemId == "" {
		return shared.ErrorResponse(400, "Missing readingLogItemId in request body")
	}

	// Find the index of the reading log item using the provided readingLogItemId.
	var indexToUpdate int
	var found bool
	for i, item := range profile.ReadingLog {
		// Assumes the log item's identifier is stored in BookID.
		if item.BookID == updateReq.ReadingLogItemId {
			indexToUpdate = i
			found = true
			break
		}
	}
	if !found {
		return shared.ErrorResponse(404, "Reading log item not found")
	}

	// Update the reading log item.
	// Here, we update PagesRead and Notes; you could also update a timestamp.
	profile.ReadingLog[indexToUpdate].PagesRead = updateReq.PagesRead
	profile.ReadingLog[indexToUpdate].Notes = updateReq.Notes
	// Optionally, update a last-updated timestamp:
	// profile.ReadingLog[indexToUpdate].LastUpdated = time.Now().Format(time.RFC3339)

	// Marshal the updated profile back into a map for DynamoDB.
	updatedProfile, err := dynamodbattribute.MarshalMap(profile)
	if err != nil {
		log.Printf("Error marshalling updated profile: %v\n", err)
		return shared.ErrorResponse(500, "Error marshalling updated profile: "+err.Error())
	}

	putInput := &dynamodb.PutItemInput{
		TableName: aws.String("PROFILES_TABLE_NAME"),
		Item:      updatedProfile,
	}
	_, err = svc.PutItem(putInput)
	if err != nil {
		log.Printf("DynamoDB PutItem error: %v\n", err)
		return shared.ErrorResponse(500, fmt.Sprintf("DynamoDB PutItem error: %v", err))
	}

	log.Printf("Reading log item updated for user %s\n", userId)
	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Body:       "Reading log item updated successfully",
	}
}

func DeleteReadingLogItem(request events.APIGatewayProxyRequest) events.APIGatewayProxyResponse {
	log.Println("DeleteReadingLog invoked")
	userId, err := shared.GetUserIDFromToken(request)
	if err != nil {
		log.Printf("Error extracting userId: %v\n", err)
		return shared.ErrorResponse(401, err.Error())
	}

	svc := shared.DynamoDBClient()
	getInput := &dynamodb.GetItemInput{
		TableName: aws.String(ProfilesTableName),
		Key: map[string]*dynamodb.AttributeValue{
			"_id": {S: aws.String(userId)},
		},
	}

	result, err := svc.GetItem(getInput)
	if err != nil {
		return shared.ErrorResponse(500, fmt.Sprintf("Error retrieving profile: %v", err))
	}
	if result.Item == nil {
		return shared.ErrorResponse(404, "Profile not found")
	}

	var profile models.Profile
	if err := dynamodbattribute.UnmarshalMap(result.Item, &profile); err != nil {
		return shared.ErrorResponse(500, fmt.Sprintf("Error unmarshalling profile: %v", err))
	}

	readingLogId, hasReadingLogId := request.QueryStringParameters["readingLogId"]
	if !hasReadingLogId || readingLogId == "" {
		return shared.ErrorResponse(400, "Missing query string parameter: readingLogId")
	}

	// Find the index of the book in the reading log
	var indexToDelete int
	var found bool
	for i, item := range profile.ReadingLog {
		if item.Id == readingLogId {
			indexToDelete = i
			found = true
			break
		}
	}
	if !found {
		return shared.ErrorResponse(404, "Entry not found in reading log")
	}

	// Remove the book from the reading log
	profile.ReadingLog = append(profile.ReadingLog[:indexToDelete], profile.ReadingLog[indexToDelete+1:]...)

	// Update the profile in DynamoDB
	updatedProfile, err := dynamodbattribute.MarshalMap(profile)
	if err != nil {
		log.Printf("Error marshalling updated profile: %v\n", err)
		return shared.ErrorResponse(500, "Error marshalling updated profile: "+err.Error())
	}

	putInput := &dynamodb.PutItemInput{
		TableName: aws.String(PROFILES_TABLE_NAME),
		Item:      updatedProfile,
	}
	_, err = svc.PutItem(putInput)
	if err != nil {
		log.Printf("DynamoDB PutItem error: %v\n", err)
		return shared.ErrorResponse(500, fmt.Sprintf("DynamoDB PutItem error: %v", err))
	}

	log.Printf("Reading log item delete for user %s\n", userId)
	return events.APIGatewayProxyResponse{
		StatusCode: 201,
		Body:       "Reading log item deleted successfully",
	}
}
