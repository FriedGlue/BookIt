package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

// We'll read this from environment variables or pass it in at deploy time
var (
	PROFILES_TABLE_NAME = os.Getenv("PROFILES_TABLE_NAME") // e.g. "UserProfiles"
)

// getUserIDFromToken extracts the user’s Cognito “sub” claim
func getUserIDFromToken(request events.APIGatewayProxyRequest) (string, error) {
	claims, ok := request.RequestContext.Authorizer["claims"].(map[string]interface{})
	if !ok {
		log.Println("No claims found in request context")
		return "", fmt.Errorf("No claims found in request context")
	}
	sub, ok := claims["sub"].(string)
	if !ok || sub == "" {
		log.Println("No 'sub' claim in token")
		return "", fmt.Errorf("No 'sub' claim in token")
	}
	return sub, nil
}

// ----------------------- Handlers -----------------------

// GetProfile retrieves the user’s profile from DynamoDB
func GetProfile(request events.APIGatewayProxyRequest) events.APIGatewayProxyResponse {
	log.Println("GetProfile invoked")
	userId, err := getUserIDFromToken(request)
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

	var profile Profile
	if err := dynamodbattribute.UnmarshalMap(result.Item, &profile); err != nil {
		log.Printf("Error unmarshalling profile: %v\n", err)
		return errorResponse(500, "Error unmarshalling profile: "+err.Error())
	}

	responseBody, err := json.Marshal(profile)
	if err != nil {
		log.Printf("Error marshalling response: %v\n", err)
		return errorResponse(500, "Error marshalling profile response")
	}

	log.Println("Profile retrieval successful")
	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Body:       string(responseBody),
	}
}

// CreateOrUpdateProfile either creates a new profile or updates an existing one
func CreateOrUpdateProfile(request events.APIGatewayProxyRequest) events.APIGatewayProxyResponse {
	log.Println("CreateOrUpdateProfile invoked")
	userId, err := getUserIDFromToken(request)
	if err != nil {
		log.Printf("Error extracting userId: %v\n", err)
		return errorResponse(401, err.Error())
	}

	var incomingProfile Profile
	if err := json.Unmarshal([]byte(request.Body), &incomingProfile); err != nil {
		log.Printf("Invalid JSON: %v\n", err)
		return errorResponse(400, "Invalid JSON: "+err.Error())
	}

	incomingProfile.ID = userId
	log.Printf("Creating/Updating profile for userId: %s\n", userId)

	item, err := dynamodbattribute.MarshalMap(incomingProfile)
	if err != nil {
		log.Printf("Error marshalling profile: %v\n", err)
		return errorResponse(500, "Error marshalling profile: "+err.Error())
	}

	svc := DynamoDBClient()
	input := &dynamodb.PutItemInput{
		TableName: aws.String(PROFILES_TABLE_NAME),
		Item:      item,
	}
	_, err = svc.PutItem(input)
	if err != nil {
		log.Printf("DynamoDB PutItem error: %v\n", err)
		return errorResponse(500, fmt.Sprintf("DynamoDB PutItem error: %v", err))
	}

	log.Printf("Profile created/updated for user %s\n", userId)
	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Body:       fmt.Sprintf("Profile created or updated for user %s", userId),
	}
}

// DeleteProfile removes a user's profile from DynamoDB
func DeleteProfile(request events.APIGatewayProxyRequest) events.APIGatewayProxyResponse {
	log.Println("DeleteProfile invoked")
	userId, err := getUserIDFromToken(request)
	if err != nil {
		log.Printf("Error extracting userId: %v\n", err)
		return errorResponse(401, err.Error())
	}

	svc := DynamoDBClient()
	input := &dynamodb.DeleteItemInput{
		TableName: aws.String(PROFILES_TABLE_NAME),
		Key: map[string]*dynamodb.AttributeValue{
			"_id": {S: aws.String(userId)},
		},
	}

	_, err = svc.DeleteItem(input)
	if err != nil {
		log.Printf("DynamoDB DeleteItem error: %v\n", err)
		return errorResponse(500, fmt.Sprintf("DynamoDB DeleteItem error: %v", err))
	}

	log.Printf("Profile deleted for user %s\n", userId)
	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Body:       fmt.Sprintf("Profile deleted for user %s", userId),
	}
}

// Define all your types (Profile, ProfileInformation, etc.) here or in another file.
type Profile struct {
	ID                 string             `json:"_id"` // Partition Key = userId
	ProfileInformation ProfileInformation `json:"profileInformation"`
	CurrentlyReading   []ReadingItem      `json:"currentlyReading,omitempty"`
	Lists              UserLists          `json:"lists,omitempty"`
	ReadingLog         []ReadingLogItem   `json:"readingLog,omitempty"`
}

type ProfileInformation struct {
	Username string `json:"username,omitempty"`
	Email    string `json:"email,omitempty"`
}

type ReadingItem struct {
	Book Book `json:"Book"`
}

type Book struct {
	BookID     string          `json:"bookId"`
	ISBN       string          `json:"isbn,omitempty"`
	Title      string          `json:"title,omitempty"`
	CoverImage string          `json:"coverImage,omitempty"`
	Progress   ReadingProgress `json:"progress,omitempty"`
}

type ReadingProgress struct {
	LastPageRead int     `json:"lastPageRead,omitempty"`
	Percentage   float64 `json:"percentage,omitempty"`
	StartedDate  string  `json:"startedDate,omitempty"`
	Notes        string  `json:"notes,omitempty"`
	LastUpdated  string  `json:"lastUpdated,omitempty"`
}

type UserLists struct {
	ToBeRead    []ToBeReadItem              `json:"toBeRead,omitempty"`
	Read        []ReadItem                  `json:"read,omitempty"`
	CustomLists map[string][]CustomListItem `json:"customLists,omitempty"`
}

type ToBeReadItem struct {
	BookID    string `json:"bookId"`
	Thumbnail string `json:"thumbnail,omitempty"`
	AddedDate string `json:"addedDate,omitempty"`
	Order     int    `json:"order,omitempty"`
}

type ReadItem struct {
	BookID        string `json:"bookId"`
	CompletedDate string `json:"completedDate,omitempty"`
	Rating        int    `json:"rating,omitempty"`
	Order         int    `json:"order,omitempty"`
	Review        string `json:"review,omitempty"`
}

type CustomListItem struct {
	BookID    string `json:"bookId"`
	Thumbnail string `json:"thumbnail,omitempty"`
	AddedDate string `json:"addedDate,omitempty"`
	Order     int    `json:"order,omitempty"`
}

type ReadingLogItem struct {
	Date             string `json:"date"`
	BookID           string `json:"bookId"`
	BookThumbnail    string `json:"bookThumbnail,omitempty"`
	PagesRead        int    `json:"pagesRead,omitempty"`
	TimeSpentMinutes int    `json:"timeSpentMinutes,omitempty"`
	Notes            string `json:"notes,omitempty"`
}
