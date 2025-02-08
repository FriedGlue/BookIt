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
type AddToBookshelfRequest struct {
	ShelfType string `json:"shelfType"` // "toBeRead", "read", or custom shelf name
	BookID    string `json:"bookId"`
	Rating    int    `json:"rating,omitempty"`    // Only for read shelf
	Review    string `json:"review,omitempty"`    // Only for read shelf
	Thumbnail string `json:"thumbnail,omitempty"` // For toBeRead and custom shelves
}

type UpdateBookshelfItemRequest struct {
	ShelfType string `json:"shelfType"`
	BookID    string `json:"bookId"`
	Rating    int    `json:"rating,omitempty"`
	Review    string `json:"review,omitempty"`
	Order     int    `json:"order,omitempty"`
}

// GetBookshelf retrieves specific bookshelves (toBeRead, read, or custom) from the Profile, or all shelves if no type is provided
func GetBookshelf(request events.APIGatewayProxyRequest) events.APIGatewayProxyResponse {
	log.Println("GetBookshelf invoked")
	userId, err := shared.GetUserIDFromToken(request)
	if err != nil {
		log.Printf("Error extracting userId: %v\n", err)
		return shared.ErrorResponse(401, err.Error())
	}

	shelfType := request.QueryStringParameters["shelfType"]

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
	if shelfType == "" {
		// If no shelfType is provided, return all shelves
		allShelves := struct {
			ToBeRead      []models.ToBeReadBook               `json:"toBeRead"`
			Read          []models.ReadBook                   `json:"read"`
			CustomShelves map[string][]models.CustomShelfBook `json:"customShelves"`
		}{
			ToBeRead:      profile.Bookshelves.ToBeRead,
			Read:          profile.Bookshelves.Read,
			CustomShelves: profile.Bookshelves.CustomShelves,
		}
		responseBody, err = json.Marshal(allShelves)
	} else {
		switch shelfType {
		case "toBeRead":
			responseBody, err = json.Marshal(profile.Bookshelves.ToBeRead)
		case "read":
			responseBody, err = json.Marshal(profile.Bookshelves.Read)
		default:
			if customShelf, exists := profile.Bookshelves.CustomShelves[shelfType]; exists {
				responseBody, err = json.Marshal(customShelf)
			} else {
				return shared.ErrorResponse(404, "Shelf not found")
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

// AddToBookshelf adds a book to a specific shelf (toBeRead, read, or custom)
func AddToBookshelf(request events.APIGatewayProxyRequest) events.APIGatewayProxyResponse {
	log.Println("AddToBookshelf invoked")
	userId, err := shared.GetUserIDFromToken(request)
	if err != nil {
		log.Printf("Error extracting userId: %v\n", err)
		return shared.ErrorResponse(401, err.Error())
	}

	var addReq AddToBookshelfRequest
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

	switch addReq.ShelfType {
	case "toBeRead":
		item := models.ToBeReadBook{
			BookID:    bookDetails.BookID,
			Thumbnail: bookDetails.CoverImageURL,
			AddedDate: currentTime,
			Title:     bookDetails.Title,
			Authors:   bookDetails.Authors,
			Order:     len(profile.Bookshelves.ToBeRead),
		}
		profile.Bookshelves.ToBeRead = append(profile.Bookshelves.ToBeRead, item)
	case "read":
		item := models.ReadBook{
			BookID:        bookDetails.BookID,
			CompletedDate: currentTime,
			Thumbnail:     bookDetails.CoverImageURL,
			Rating:        addReq.Rating,
			Review:        addReq.Review,
			Title:         bookDetails.Title,
			Authors:       bookDetails.Authors,
			Order:         len(profile.Bookshelves.Read),
		}
		profile.Bookshelves.Read = append(profile.Bookshelves.Read, item)
	default:
		item := models.CustomShelfBook{
			BookID:    bookDetails.BookID,
			Thumbnail: bookDetails.CoverImageURL,
			AddedDate: currentTime,
			Title:     bookDetails.Title,
			Authors:   bookDetails.Authors,
			Order:     len(profile.Bookshelves.CustomShelves[addReq.ShelfType]),
		}
		if profile.Bookshelves.CustomShelves == nil {
			profile.Bookshelves.CustomShelves = make(map[string][]models.CustomShelfBook)
		}
		profile.Bookshelves.CustomShelves[addReq.ShelfType] = append(profile.Bookshelves.CustomShelves[addReq.ShelfType], item)
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
		Body:       "Book added to shelf successfully",
	}
}

// DeleteBookshelf deletes a custom shelf from a user's profile
func DeleteBookshelf(request events.APIGatewayProxyRequest) events.APIGatewayProxyResponse {
	log.Println("DeleteBookshelf invoked")
	userId, err := shared.GetUserIDFromToken(request)
	if err != nil {
		log.Printf("Error extracting userId: %v\n", err)
		return shared.ErrorResponse(401, err.Error())
	}

	shelfName := request.QueryStringParameters["shelfName"]
	if shelfName == "" {
		return shared.ErrorResponse(400, "shelfName parameter is required")
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

	if _, exists := profile.Bookshelves.CustomShelves[shelfName]; !exists {
		return shared.ErrorResponse(404, "Shelf not found")
	}

	delete(profile.Bookshelves.CustomShelves, shelfName)

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
		Body:       "Shelf deleted successfully",
	}
}

// RemoveFromBookshelf removes a book from a specific shelf
func DeleteBookshelfItem(request events.APIGatewayProxyRequest) events.APIGatewayProxyResponse {
	log.Println("RemoveFromBookshelf invoked")
	userId, err := shared.GetUserIDFromToken(request)
	if err != nil {
		log.Printf("Error extracting userId: %v\n", err)
		return shared.ErrorResponse(401, err.Error())
	}

	shelfType := request.QueryStringParameters["shelfType"]
	bookId := request.QueryStringParameters["bookId"]
	if shelfType == "" || bookId == "" {
		return shared.ErrorResponse(400, "shelfType and bookId parameters are required")
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
	switch shelfType {
	case "toBeRead":
		for i, item := range profile.Bookshelves.ToBeRead {
			if item.BookID == bookId {
				profile.Bookshelves.ToBeRead = append(profile.Bookshelves.ToBeRead[:i], profile.Bookshelves.ToBeRead[i+1:]...)
				found = true
				break
			}
		}
	case "read":
		for i, item := range profile.Bookshelves.Read {
			if item.BookID == bookId {
				profile.Bookshelves.Read = append(profile.Bookshelves.Read[:i], profile.Bookshelves.Read[i+1:]...)
				found = true
				break
			}
		}
	default:
		if customShelf, exists := profile.Bookshelves.CustomShelves[shelfType]; exists {
			for i, item := range customShelf {
				if item.BookID == bookId {
					profile.Bookshelves.CustomShelves[shelfType] = append(customShelf[:i], customShelf[i+1:]...)
					found = true
					break
				}
			}
		}
	}

	if !found {
		return shared.ErrorResponse(404, "Book not found in the specified shelf")
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
		Body:       "Book removed from shelf successfully",
	}
}

// CreateBookshelf creates a new custom shelf for a user
func CreateBookshelf(request events.APIGatewayProxyRequest) events.APIGatewayProxyResponse {
	log.Println("CreateBookshelf invoked")
	userId, err := shared.GetUserIDFromToken(request)
	if err != nil {
		log.Printf("Error extracting userId: %v\n", err)
		return shared.ErrorResponse(401, err.Error())
	}

	var input struct {
		ShelfName string `json:"shelfName"`
	}
	if err := json.Unmarshal([]byte(request.Body), &input); err != nil {
		log.Printf("Invalid JSON: %v\n", err)
		return shared.ErrorResponse(400, "Invalid JSON: "+err.Error())
	}
	if input.ShelfName == "" {
		return shared.ErrorResponse(400, "shelfName is required")
	}

	svc := shared.DynamoDBClient()
	getInput := &dynamodb.GetItemInput{
		TableName: aws.String(PROFILES_TABLE_NAME),
		Key: map[string]*dynamodb.AttributeValue{
			"_id": {S: aws.String(userId)},
		},
	}

	result, err := svc.GetItem(getInput)
	if err != nil {
		log.Printf("DynamoDB GetItem error: %v\n", err)
		return shared.ErrorResponse(500, fmt.Sprintf("DynamoDB GetItem error: %v", err))
	}

	var profile models.Profile
	if result.Item != nil {
		if err := dynamodbattribute.UnmarshalMap(result.Item, &profile); err != nil {
			log.Printf("Error unmarshalling profile: %v\n", err)
			return shared.ErrorResponse(500, "Error unmarshalling profile: "+err.Error())
		}
	} else {
		// Initialize a new profile if it doesn't exist
		profile = models.Profile{
			ID: userId,
		}
	}

	// Initialize CustomShelves if it doesn't exist
	if profile.Bookshelves.CustomShelves == nil {
		profile.Bookshelves.CustomShelves = make(map[string][]models.CustomShelfBook)
	}

	// Check if the shelf already exists
	if _, exists := profile.Bookshelves.CustomShelves[input.ShelfName]; exists {
		return shared.ErrorResponse(400, "A shelf with this name already exists")
	}

	// Create the new shelf
	profile.Bookshelves.CustomShelves[input.ShelfName] = []models.CustomShelfBook{}

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
		Body:       fmt.Sprintf("Custom shelf '%s' created successfully", input.ShelfName),
	}
}
