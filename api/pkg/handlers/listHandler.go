package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/FriedGlue/BookIt/api/pkg/models"
	"github.com/FriedGlue/BookIt/api/pkg/shared"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

// Request structs
type AddToListRequest struct {
	ListType  string `json:"listType"` // "toBeRead", "read", or custom list name
	BookID    string `json:"bookId"`
	Rating    int    `json:"rating,omitempty"`    // Only for read list
	Review    string `json:"review,omitempty"`    // Only for read list
	Thumbnail string `json:"thumbnail,omitempty"` // For toBeRead and custom lists
}

type UpdateListItemRequest struct {
	ListType string `json:"listType"`
	BookID   string `json:"bookId"`
	Rating   int    `json:"rating,omitempty"`
	Review   string `json:"review,omitempty"`
	Order    int    `json:"order,omitempty"`
}

// GetList retrieves specific lists (toBeRead, read, or custom) from the Profile, or all lists if no type is provided
func GetList(request events.APIGatewayProxyRequest) events.APIGatewayProxyResponse {
	log.Println("GetList invoked")
	userId, err := shared.GetUserIDFromToken(request)
	if err != nil {
		log.Printf("Error extracting userId: %v\n", err)
		return shared.ErrorResponse(401, err.Error())
	}

	listType := request.QueryStringParameters["listType"]

	svc := shared.DynamoDBClient()
	input := &dynamodb.GetItemInput{
		TableName: aws.String(PROFILES_TABLE_NAME),
		Key: map[string]*dynamodb.AttributeValue{
			"_id": {S: aws.String(userId)},
		},
	}

	result, err := svc.GetItem(input)
	if err != nil {
		log.Printf("DynamoDB GetItem error: %v\n", err)
		return shared.ErrorResponse(500, fmt.Sprintf("DynamoDB GetItem error: %v", err))
	}
	if result.Item == nil {
		return shared.ErrorResponse(404, "Profile not found")
	}

	var profile models.Profile
	if err := dynamodbattribute.UnmarshalMap(result.Item, &profile); err != nil {
		log.Printf("Error unmarshalling profile: %v\n", err)
		return shared.ErrorResponse(500, "Error unmarshalling profile: "+err.Error())
	}

	var responseBody []byte
	if listType == "" {
		// If no listType is provided, return all lists
		allLists := struct {
			ToBeRead []models.ToBeReadItem              `json:"toBeRead"`
			Read     []models.ReadItem                  `json:"read"`
			Custom   map[string][]models.CustomListItem `json:"customLists"`
		}{
			ToBeRead: profile.Lists.ToBeRead,
			Read:     profile.Lists.Read,
			Custom:   profile.Lists.CustomLists,
		}
		responseBody, err = json.Marshal(allLists)
	} else {
		switch listType {
		case "toBeRead":
			responseBody, err = json.Marshal(profile.Lists.ToBeRead)
		case "read":
			responseBody, err = json.Marshal(profile.Lists.Read)
		default:
			if customList, exists := profile.Lists.CustomLists[listType]; exists {
				responseBody, err = json.Marshal(customList)
			} else {
				return shared.ErrorResponse(404, "List not found")
			}
		}
	}

	if err != nil {
		log.Printf("Error marshalling response: %v\n", err)
		return shared.ErrorResponse(500, "Error marshalling response")
	}

	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Body:       string(responseBody),
	}
}

// AddToList adds a book to a specific list (toBeRead, read, or custom)
func AddToList(request events.APIGatewayProxyRequest) events.APIGatewayProxyResponse {
	log.Println("AddToList invoked")
	userId, err := shared.GetUserIDFromToken(request)
	if err != nil {
		log.Printf("Error extracting userId: %v\n", err)
		return shared.ErrorResponse(401, err.Error())
	}

	var addReq AddToListRequest
	if err := json.Unmarshal([]byte(request.Body), &addReq); err != nil {
		log.Printf("Invalid JSON: %v\n", err)
		return shared.ErrorResponse(400, "Invalid JSON: "+err.Error())
	}

	// First get the book details from books table
	svc := shared.DynamoDBClient()
	bookInput := &dynamodb.GetItemInput{
		TableName: aws.String(BOOKS_TABLE_NAME),
		Key: map[string]*dynamodb.AttributeValue{
			"bookId": {S: aws.String(addReq.BookID)},
		},
	}

	bookResult, err := svc.GetItem(bookInput)
	if err != nil {
		log.Printf("DynamoDB GetItem error for book: %v\n", err)
		return shared.ErrorResponse(500, fmt.Sprintf("DynamoDB GetItem error: %v", err))
	}
	if bookResult.Item == nil {
		log.Println("Book not found")
		return shared.ErrorResponse(404, "Book not found")
	}

	var bookDetails BookData
	if err := dynamodbattribute.UnmarshalMap(bookResult.Item, &bookDetails); err != nil {
		log.Printf("Error unmarshalling book details: %v\n", err)
		return shared.ErrorResponse(500, "Error unmarshalling book details: "+err.Error())
	}

	// Get user profile
	profileInput := &dynamodb.GetItemInput{
		TableName: aws.String(PROFILES_TABLE_NAME),
		Key: map[string]*dynamodb.AttributeValue{
			"_id": {S: aws.String(userId)},
		},
	}

	result, err := svc.GetItem(profileInput)
	if err != nil {
		log.Printf("DynamoDB GetItem error: %v\n", err)
		return shared.ErrorResponse(500, fmt.Sprintf("DynamoDB GetItem error: %v", err))
	}
	if result.Item == nil {
		return shared.ErrorResponse(404, "Profile not found")
	}

	var profile models.Profile
	if err := dynamodbattribute.UnmarshalMap(result.Item, &profile); err != nil {
		log.Printf("Error unmarshalling profile: %v\n", err)
		return shared.ErrorResponse(500, "Error unmarshalling profile: "+err.Error())
	}

	currentTime := time.Now().Format(time.RFC3339)

	switch addReq.ListType {
	case "toBeRead":
		item := models.ToBeReadItem{
			BookID:    bookDetails.BookID,
			Thumbnail: bookDetails.CoverImageURL,
			AddedDate: currentTime,
			Title:     bookDetails.Title,
			Authors:   bookDetails.Authors,
			Order:     len(profile.Lists.ToBeRead),
		}
		profile.Lists.ToBeRead = append(profile.Lists.ToBeRead, item)
	case "read":
		item := models.ReadItem{
			BookID:        bookDetails.BookID,
			CompletedDate: currentTime,
			Thumbnail:     bookDetails.CoverImageURL,
			Rating:        addReq.Rating,
			Review:        addReq.Review,
			Title:         bookDetails.Title,
			Authors:       bookDetails.Authors,
			Order:         len(profile.Lists.Read),
		}
		profile.Lists.Read = append(profile.Lists.Read, item)
	default:
		item := models.CustomListItem{
			BookID:    bookDetails.BookID,
			Thumbnail: bookDetails.CoverImageURL,
			AddedDate: currentTime,
			Title:     bookDetails.Title,
			Authors:   bookDetails.Authors,
			Order:     len(profile.Lists.CustomLists[addReq.ListType]),
		}
		if profile.Lists.CustomLists == nil {
			profile.Lists.CustomLists = make(map[string][]models.CustomListItem)
		}
		profile.Lists.CustomLists[addReq.ListType] = append(profile.Lists.CustomLists[addReq.ListType], item)
	}

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

	return events.APIGatewayProxyResponse{
		StatusCode: 201,
		Body:       "Book added to list successfully",
	}
}

// UpdateListItem updates an item in a specific list
func UpdateListItem(request events.APIGatewayProxyRequest) events.APIGatewayProxyResponse {
	log.Println("UpdateListItem invoked")
	userId, err := shared.GetUserIDFromToken(request)
	if err != nil {
		log.Printf("Error extracting userId: %v\n", err)
		return shared.ErrorResponse(401, err.Error())
	}

	var updateReq UpdateListItemRequest
	if err := json.Unmarshal([]byte(request.Body), &updateReq); err != nil {
		log.Printf("Invalid JSON: %v\n", err)
		return shared.ErrorResponse(400, "Invalid JSON: "+err.Error())
	}

	svc := shared.DynamoDBClient()
	input := &dynamodb.GetItemInput{
		TableName: aws.String(PROFILES_TABLE_NAME),
		Key: map[string]*dynamodb.AttributeValue{
			"_id": {S: aws.String(userId)},
		},
	}

	result, err := svc.GetItem(input)
	if err != nil {
		log.Printf("DynamoDB GetItem error: %v\n", err)
		return shared.ErrorResponse(500, fmt.Sprintf("DynamoDB GetItem error: %v", err))
	}
	if result.Item == nil {
		return shared.ErrorResponse(404, "Profile not found")
	}

	var profile models.Profile
	if err := dynamodbattribute.UnmarshalMap(result.Item, &profile); err != nil {
		log.Printf("Error unmarshalling profile: %v\n", err)
		return shared.ErrorResponse(500, "Error unmarshalling profile: "+err.Error())
	}

	found := false
	switch updateReq.ListType {
	case "toBeRead":
		for i := range profile.Lists.ToBeRead {
			if profile.Lists.ToBeRead[i].BookID == updateReq.BookID {
				if updateReq.Order >= 0 {
					profile.Lists.ToBeRead[i].Order = updateReq.Order
				}
				found = true
				break
			}
		}
	case "read":
		for i := range profile.Lists.Read {
			if profile.Lists.Read[i].BookID == updateReq.BookID {
				if updateReq.Rating >= 0 {
					profile.Lists.Read[i].Rating = updateReq.Rating
				}
				if updateReq.Review != "" {
					profile.Lists.Read[i].Review = updateReq.Review
				}
				if updateReq.Order >= 0 {
					profile.Lists.Read[i].Order = updateReq.Order
				}
				found = true
				break
			}
		}
	default:
		if customList, exists := profile.Lists.CustomLists[updateReq.ListType]; exists {
			for i := range customList {
				if customList[i].BookID == updateReq.BookID {
					if updateReq.Order >= 0 {
						customList[i].Order = updateReq.Order
					}
					profile.Lists.CustomLists[updateReq.ListType] = customList
					found = true
					break
				}
			}
		}
	}

	if !found {
		return shared.ErrorResponse(404, "Book not found in the specified list")
	}

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

	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Body:       "List item updated successfully",
	}
}

// DeleteList deletes a custom list from a user's profile
func DeleteList(request events.APIGatewayProxyRequest) events.APIGatewayProxyResponse {
	log.Println("DeleteList invoked")
	userId, err := shared.GetUserIDFromToken(request)
	if err != nil {
		log.Printf("Error extracting userId: %v\n", err)
		return shared.ErrorResponse(401, err.Error())
	}

	listName := request.QueryStringParameters["listName"]
	if listName == "" {
		return shared.ErrorResponse(400, "listName parameter is required")
	}

	svc := shared.DynamoDBClient()
	input := &dynamodb.GetItemInput{
		TableName: aws.String(PROFILES_TABLE_NAME),
		Key: map[string]*dynamodb.AttributeValue{
			"_id": {S: aws.String(userId)},
		},
	}

	result, err := svc.GetItem(input)
	if err != nil {
		log.Printf("DynamoDB GetItem error: %v\n", err)
		return shared.ErrorResponse(500, fmt.Sprintf("DynamoDB GetItem error: %v", err))
	}
	if result.Item == nil {
		return shared.ErrorResponse(404, "Profile not found")
	}

	var profile models.Profile
	if err := dynamodbattribute.UnmarshalMap(result.Item, &profile); err != nil {
		log.Printf("Error unmarshalling profile: %v\n", err)
		return shared.ErrorResponse(500, "Error unmarshalling profile: "+err.Error())
	}

	if _, exists := profile.Lists.CustomLists[listName]; !exists {
		return shared.ErrorResponse(404, "List not found")
	}

	delete(profile.Lists.CustomLists, listName)

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

	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Body:       "List deleted successfully",
	}
}

// RemoveFromList removes a book from a specific list
func DeleteListItem(request events.APIGatewayProxyRequest) events.APIGatewayProxyResponse {
	log.Println("RemoveFromList invoked")
	userId, err := shared.GetUserIDFromToken(request)
	if err != nil {
		log.Printf("Error extracting userId: %v\n", err)
		return shared.ErrorResponse(401, err.Error())
	}

	listType := request.QueryStringParameters["listType"]
	bookId := request.QueryStringParameters["bookId"]
	if listType == "" || bookId == "" {
		return shared.ErrorResponse(400, "listType and bookId parameters are required")
	}

	svc := shared.DynamoDBClient()
	input := &dynamodb.GetItemInput{
		TableName: aws.String(PROFILES_TABLE_NAME),
		Key: map[string]*dynamodb.AttributeValue{
			"_id": {S: aws.String(userId)},
		},
	}

	result, err := svc.GetItem(input)
	if err != nil {
		log.Printf("DynamoDB GetItem error: %v\n", err)
		return shared.ErrorResponse(500, fmt.Sprintf("DynamoDB GetItem error: %v", err))
	}
	if result.Item == nil {
		return shared.ErrorResponse(404, "Profile not found")
	}

	var profile models.Profile
	if err := dynamodbattribute.UnmarshalMap(result.Item, &profile); err != nil {
		log.Printf("Error unmarshalling profile: %v\n", err)
		return shared.ErrorResponse(500, "Error unmarshalling profile: "+err.Error())
	}

	// Initialize lists if they don't exist

	found := false
	switch listType {
	case "toBeRead":
		for i, item := range profile.Lists.ToBeRead {
			if item.BookID == bookId {
				profile.Lists.ToBeRead = append(profile.Lists.ToBeRead[:i], profile.Lists.ToBeRead[i+1:]...)
				found = true
				break
			}
		}
	case "read":
		for i, item := range profile.Lists.Read {
			if item.BookID == bookId {
				profile.Lists.Read = append(profile.Lists.Read[:i], profile.Lists.Read[i+1:]...)
				found = true
				break
			}
		}
	default:
		if customList, exists := profile.Lists.CustomLists[listType]; exists {
			for i, item := range customList {
				if item.BookID == bookId {
					profile.Lists.CustomLists[listType] = append(customList[:i], customList[i+1:]...)
					found = true
					break
				}
			}
		}
	}

	if !found {
		return shared.ErrorResponse(404, "Book not found in the specified list")
	}

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

	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Body:       "Book removed from list successfully",
	}
}
