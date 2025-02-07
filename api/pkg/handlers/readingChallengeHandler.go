package handlers

import (
	"encoding/json"
	"log"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/google/uuid"

	"github.com/FriedGlue/BookIt/api/pkg/models"
	"github.com/FriedGlue/BookIt/api/pkg/shared"
)

// CreateChallenge creates a new reading challenge and appends it to the profile's Challenges field.
func CreateChallenge(request events.APIGatewayProxyRequest) events.APIGatewayProxyResponse {
	log.Println("CreateChallenge invoked")

	var challenge models.ReadingChallenge
	if err := json.Unmarshal([]byte(request.Body), &challenge); err != nil {
		return shared.ErrorResponse(400, "Invalid request body")
	}

	// Get user ID from token
	userID, err := shared.GetUserIDFromToken(request)
	if err != nil {
		return shared.ErrorResponse(401, err.Error())
	}

	// Initialize challenge fields
	challenge.ID = uuid.New().String()
	challenge.UserID = userID
	challenge.CreatedAt = time.Now()
	challenge.UpdatedAt = time.Now()

	// Calculate required rate and initialize progress
	requiredRate, unit := calculateRequiredRate(challenge)
	challenge.Progress = models.ChallengeProgress{
		Current:    0,
		Percentage: 0,
		Rate: struct {
			Current  float64 `json:"current"`
			Required float64 `json:"required"`
			Unit     string  `json:"unit"`
		}{
			Current:  0,
			Required: requiredRate,
			Unit:     unit,
		},
	}

	// Fetch the user's profile from DynamoDB using the user ID (stored as "_id")
	svc := shared.DynamoDBClient()
	getInput := &dynamodb.GetItemInput{
		TableName: aws.String(ProfilesTableName),
		Key: map[string]*dynamodb.AttributeValue{
			"_id": {S: aws.String(userID)},
		},
	}
	result, err := svc.GetItem(getInput)
	if err != nil {
		return shared.ErrorResponse(500, "Error fetching profile")
	}
	if result.Item == nil {
		return shared.ErrorResponse(404, "Profile not found")
	}

	var profile models.Profile
	if err := dynamodbattribute.UnmarshalMap(result.Item, &profile); err != nil {
		return shared.ErrorResponse(500, "Error unmarshalling profile")
	}

	// Append the new challenge to the profile's Challenges slice
	profile.Challenges = append(profile.Challenges, challenge)

	// Marshal the updated profile back to a map and write it back to DynamoDB
	updatedProfile, err := dynamodbattribute.MarshalMap(profile)
	if err != nil {
		return shared.ErrorResponse(500, "Error marshalling updated profile")
	}

	_, err = svc.PutItem(&dynamodb.PutItemInput{
		TableName: aws.String(ProfilesTableName),
		Item:      updatedProfile,
	})
	if err != nil {
		return shared.ErrorResponse(500, "Error saving updated profile")
	}

	return shared.SuccessResponse(201, challenge)
}

// GetChallenges retrieves all reading challenges from the profile.
func GetChallenges(request events.APIGatewayProxyRequest) events.APIGatewayProxyResponse {
	log.Println("GetChallenges invoked")

	userID, err := shared.GetUserIDFromToken(request)
	if err != nil {
		return shared.ErrorResponse(401, err.Error())
	}

	svc := shared.DynamoDBClient()
	getInput := &dynamodb.GetItemInput{
		TableName: aws.String(ProfilesTableName),
		Key: map[string]*dynamodb.AttributeValue{
			"_id": {S: aws.String(userID)},
		},
	}
	result, err := svc.GetItem(getInput)
	if err != nil {
		return shared.ErrorResponse(500, "Error fetching profile")
	}
	if result.Item == nil {
		return shared.ErrorResponse(404, "Profile not found")
	}

	var profile models.Profile
	if err := dynamodbattribute.UnmarshalMap(result.Item, &profile); err != nil {
		return shared.ErrorResponse(500, "Error unmarshalling profile")
	}

	// Return the challenges slice from the profile
	return shared.SuccessResponse(200, profile.Challenges)
}

// UpdateChallenge updates a specific reading challenge within the profile.
func UpdateChallenge(request events.APIGatewayProxyRequest) events.APIGatewayProxyResponse {
	log.Println("UpdateChallenge invoked")

	challengeID := request.PathParameters["id"]
	userID, err := shared.GetUserIDFromToken(request)
	if err != nil {
		return shared.ErrorResponse(401, err.Error())
	}

	// Define the update payload structure
	var updateData struct {
		Current int `json:"current"`
	}
	if err := json.Unmarshal([]byte(request.Body), &updateData); err != nil {
		return shared.ErrorResponse(400, "Invalid request body")
	}

	svc := shared.DynamoDBClient()
	// Retrieve the user's profile
	getInput := &dynamodb.GetItemInput{
		TableName: aws.String(ProfilesTableName),
		Key: map[string]*dynamodb.AttributeValue{
			"_id": {S: aws.String(userID)},
		},
	}
	result, err := svc.GetItem(getInput)
	if err != nil {
		return shared.ErrorResponse(500, "Error fetching profile")
	}
	if result.Item == nil {
		return shared.ErrorResponse(404, "Profile not found")
	}

	var profile models.Profile
	if err := dynamodbattribute.UnmarshalMap(result.Item, &profile); err != nil {
		return shared.ErrorResponse(500, "Error unmarshalling profile")
	}

	// Locate and update the specified challenge in the profile
	found := false
	for i, ch := range profile.Challenges {
		if ch.ID == challengeID {
			profile.Challenges[i].Progress.Current = updateData.Current
			if profile.Challenges[i].Target != 0 {
				profile.Challenges[i].Progress.Percentage = float64(updateData.Current) / float64(profile.Challenges[i].Target) * 100
			}
			profile.Challenges[i].Progress.Rate.Current = calculateCurrentRate(profile.Challenges[i])
			profile.Challenges[i].UpdatedAt = time.Now()
			found = true
			break
		}
	}
	if !found {
		return shared.ErrorResponse(404, "Challenge not found")
	}

	// Write the updated profile back to DynamoDB
	updatedProfile, err := dynamodbattribute.MarshalMap(profile)
	if err != nil {
		return shared.ErrorResponse(500, "Error marshalling updated profile")
	}
	_, err = svc.PutItem(&dynamodb.PutItemInput{
		TableName: aws.String(ProfilesTableName),
		Item:      updatedProfile,
	})
	if err != nil {
		return shared.ErrorResponse(500, "Error saving updated profile")
	}

	return shared.SuccessResponse(200, profile.Challenges)
}

// DeleteChallenge deletes a specific reading challenge from the profile.
func DeleteChallenge(request events.APIGatewayProxyRequest) events.APIGatewayProxyResponse {
	log.Println("DeleteChallenge invoked")

	challengeID := request.PathParameters["id"]
	userID, err := shared.GetUserIDFromToken(request)
	if err != nil {
		return shared.ErrorResponse(401, err.Error())
	}

	svc := shared.DynamoDBClient()
	// Retrieve the user's profile
	getInput := &dynamodb.GetItemInput{
		TableName: aws.String(ProfilesTableName),
		Key: map[string]*dynamodb.AttributeValue{
			"_id": {S: aws.String(userID)},
		},
	}
	result, err := svc.GetItem(getInput)
	if err != nil {
		return shared.ErrorResponse(500, "Error fetching profile")
	}
	if result.Item == nil {
		return shared.ErrorResponse(404, "Profile not found")
	}

	var profile models.Profile
	if err := dynamodbattribute.UnmarshalMap(result.Item, &profile); err != nil {
		return shared.ErrorResponse(500, "Error unmarshalling profile")
	}

	// Find the challenge index to delete
	indexToDelete := -1
	for i, ch := range profile.Challenges {
		if ch.ID == challengeID {
			indexToDelete = i
			break
		}
	}
	if indexToDelete < 0 {
		return shared.ErrorResponse(404, "Challenge not found")
	}

	// Remove the challenge from the slice
	profile.Challenges = append(profile.Challenges[:indexToDelete], profile.Challenges[indexToDelete+1:]...)

	// Write the updated profile back to DynamoDB
	updatedProfile, err := dynamodbattribute.MarshalMap(profile)
	if err != nil {
		return shared.ErrorResponse(500, "Error marshalling updated profile")
	}
	_, err = svc.PutItem(&dynamodb.PutItemInput{
		TableName: aws.String(ProfilesTableName),
		Item:      updatedProfile,
	})
	if err != nil {
		return shared.ErrorResponse(500, "Error deleting challenge from profile")
	}

	return shared.SuccessResponse(200, map[string]string{"message": "Challenge deleted successfully"})
}

// calculateRequiredRate computes the required reading rate based on the challenge's timeframe.
func calculateRequiredRate(challenge models.ReadingChallenge) (float64, string) {
	duration := challenge.EndDate.Sub(challenge.StartDate)
	var rate float64
	var unit string

	switch challenge.TimeFrame {
	case models.YearTimeFrame:
		monthsTotal := float64(duration.Hours()) / (24 * 30)
		rate = float64(challenge.Target) / monthsTotal
		if challenge.Type == models.BooksChallenge {
			unit = "books/month"
		} else {
			unit = "pages/month"
		}
	case models.MonthTimeFrame:
		weeksTotal := float64(duration.Hours()) / (24 * 7)
		rate = float64(challenge.Target) / weeksTotal
		if challenge.Type == models.BooksChallenge {
			unit = "books/week"
		} else {
			unit = "pages/week"
		}
	case models.WeekTimeFrame:
		daysTotal := float64(duration.Hours()) / 24
		rate = float64(challenge.Target) / daysTotal
		if challenge.Type == models.BooksChallenge {
			unit = "books/day"
		} else {
			unit = "pages/day"
		}
	}

	return rate, unit
}

// calculateCurrentRate computes the current reading rate based on progress so far.
func calculateCurrentRate(challenge models.ReadingChallenge) float64 {
	duration := time.Now().Sub(challenge.StartDate)
	var divisor float64

	switch challenge.TimeFrame {
	case models.YearTimeFrame:
		divisor = float64(duration.Hours()) / (24 * 30) // months
	case models.MonthTimeFrame:
		divisor = float64(duration.Hours()) / (24 * 7) // weeks
	case models.WeekTimeFrame:
		divisor = float64(duration.Hours()) / 24 // days
	}

	if divisor == 0 {
		return 0
	}
	return float64(challenge.Progress.Current) / divisor
}
