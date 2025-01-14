package handlers

import (
	"encoding/json"
	"fmt"
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
		return "", fmt.Errorf("No claims found in request context")
	}
	sub, ok := claims["sub"].(string)
	if !ok || sub == "" {
		return "", fmt.Errorf("No 'sub' claim in token")
	}
	return sub, nil
}

// ----------------------- Data Model -----------------------

// Profile represents the entire user profile in DynamoDB
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

// ----------------------- Handlers -----------------------

// GetProfile retrieves the user’s profile from DynamoDB
func GetProfile(request events.APIGatewayProxyRequest) events.APIGatewayProxyResponse {
	userId, err := getUserIDFromToken(request)
	if err != nil {
		return errorResponse(401, err.Error())
	}

	svc := DynamoDBClient()

	// Fetch single item by userId
	input := &dynamodb.GetItemInput{
		TableName: &PROFILES_TABLE_NAME,
		Key: map[string]*dynamodb.AttributeValue{
			"_id": {S: aws.String(userId)}, // _id is the partition key
		},
	}

	result, err := svc.GetItem(input)
	if err != nil {
		return errorResponse(500, fmt.Sprintf("DynamoDB GetItem error: %v", err))
	}
	if result.Item == nil {
		return errorResponse(404, "Profile not found")
	}

	var profile Profile
	if err := dynamodbattribute.UnmarshalMap(result.Item, &profile); err != nil {
		return errorResponse(500, "Error unmarshalling profile: "+err.Error())
	}

	responseBody, _ := json.Marshal(profile)
	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Body:       string(responseBody),
	}
}

// CreateOrUpdateProfile either creates a new profile or updates an existing one
// This can be a PUT or POST route, depending on your preference
func CreateOrUpdateProfile(request events.APIGatewayProxyRequest) events.APIGatewayProxyResponse {
	userId, err := getUserIDFromToken(request)
	if err != nil {
		return errorResponse(401, err.Error())
	}

	// Parse the incoming JSON as a full Profile structure
	var incomingProfile Profile
	if err := json.Unmarshal([]byte(request.Body), &incomingProfile); err != nil {
		return errorResponse(400, "Invalid JSON: "+err.Error())
	}

	// Ensure the _id matches the user’s sub from the token
	incomingProfile.ID = userId

	// If you want to ensure a timestamp or something else, you can do it here
	// e.g., if you store "lastUpdated" at the root

	item, err := dynamodbattribute.MarshalMap(incomingProfile)
	if err != nil {
		return errorResponse(500, "Error marshalling profile: "+err.Error())
	}

	svc := DynamoDBClient()
	input := &dynamodb.PutItemInput{
		TableName: &PROFILES_TABLE_NAME,
		Item:      item,
	}
	_, err = svc.PutItem(input)
	if err != nil {
		return errorResponse(500, fmt.Sprintf("DynamoDB PutItem error: %v", err))
	}

	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Body:       fmt.Sprintf("Profile created or updated for user %s", userId),
	}
}

// Optionally, you could have a "DeleteProfile" if you want
// The data model might not require that, but here's how you'd do it:
func DeleteProfile(request events.APIGatewayProxyRequest) events.APIGatewayProxyResponse {
	userId, err := getUserIDFromToken(request)
	if err != nil {
		return errorResponse(401, err.Error())
	}

	svc := DynamoDBClient()
	input := &dynamodb.DeleteItemInput{
		TableName: &PROFILES_TABLE_NAME,
		Key: map[string]*dynamodb.AttributeValue{
			"_id": {S: aws.String(userId)},
		},
	}
	if _, err := svc.DeleteItem(input); err != nil {
		return errorResponse(500, fmt.Sprintf("DynamoDB DeleteItem error: %v", err))
	}

	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Body:       fmt.Sprintf("Profile deleted for user %s", userId),
	}
}
