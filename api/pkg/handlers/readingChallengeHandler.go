package handlers

import (
	"encoding/json"
	"log"
	"math"
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

	// Calculate required rate and initialize progress.
	requiredRate, unit := calculateRequiredRate(challenge)
	challenge.Progress = models.ChallengeProgress{
		Current:    0,
		Percentage: 0,
		Rate: struct {
			Required     float64 `json:"required"`
			CurrentPace  float64 `json:"currentPace"`
			ScheduleDiff float64 `json:"scheduleDiff"`
			Unit         string  `json:"unit"`
			Status       string  `json:"status"`
		}{
			Required:     requiredRate,
			CurrentPace:  0,
			ScheduleDiff: 0,
			Unit:         unit,
			// Default to ON_TRACK initially.
			Status: "ON_TRACK",
		},
	}

	// If the challenge start date is in the past and no progress has been made,
	// update the schedule status accordingly.
	if time.Now().After(challenge.StartDate) && challenge.Progress.Current == 0 {
		scheduleDiff, status := calculateScheduleStatus(challenge)
		challenge.Progress.Rate.ScheduleDiff = scheduleDiff
		challenge.Progress.Rate.Status = status
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

// calculateRequiredRate computes the required reading rate based on the challenge's timeframe.
func calculateRequiredRate(challenge models.ReadingChallenge) (float64, string) {
	duration := challenge.EndDate.Sub(challenge.StartDate)
	var rate float64
	var unit string

	switch challenge.TimeFrame {
	case models.YearTimeFrame:
		monthsTotal := float64(duration.Hours()) / (24 * 30)
		rate = math.Round(float64(challenge.Target)/monthsTotal*100) / 100
		if challenge.Type == models.BooksChallenge {
			unit = "books/month"
		} else {
			unit = "pages/month"
		}
	case models.MonthTimeFrame:
		weeksTotal := float64(duration.Hours()) / (24 * 7)
		rate = math.Round(float64(challenge.Target)/weeksTotal*100) / 100
		if challenge.Type == models.BooksChallenge {
			unit = "books/week"
		} else {
			unit = "pages/week"
		}
	case models.WeekTimeFrame:
		daysTotal := float64(duration.Hours()) / 24
		rate = math.Round(float64(challenge.Target)/daysTotal*100) / 100
		if challenge.Type == models.BooksChallenge {
			unit = "books/day"
		} else {
			unit = "pages/day"
		}
	}

	return rate, unit
}

// calculateCurrentPace computes the actual reading pace
func calculateCurrentPace(challenge models.ReadingChallenge) float64 {
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

	return math.Round(float64(challenge.Progress.Current)/divisor*100) / 100
}

// calculateScheduleStatus determines if ahead/behind and by how much.
func calculateScheduleStatus(challenge models.ReadingChallenge) (float64, string) {
	duration := time.Now().Sub(challenge.StartDate)
	totalDuration := challenge.EndDate.Sub(challenge.StartDate)

	// Calculate expected progress at this point.
	expectedProgress := float64(challenge.Target) * (float64(duration) / float64(totalDuration))
	actualProgress := float64(challenge.Progress.Current)

	// If no progress and at least one minute has elapsed since the challenge started,
	// mark the challenge as behind schedule.
	if actualProgress == 0 && time.Since(challenge.StartDate) > time.Minute {
		return expectedProgress, "BEHIND"
	}

	// Calculate the progress difference.
	progressDiff := actualProgress - expectedProgress

	// If the difference is negligible, consider it on track.
	if math.Abs(progressDiff) < 0.01 {
		return 0, "ON_TRACK"
	}

	// Return the absolute difference and status.
	if progressDiff > 0 {
		return progressDiff, "AHEAD"
	}
	return math.Abs(progressDiff), "BEHIND"
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

			// Calculate and update current reading pace
			profile.Challenges[i].Progress.Rate.CurrentPace = calculateCurrentPace(profile.Challenges[i])

			// Calculate cumulative schedule difference and status,
			// then store in the new ScheduleDiff field.
			scheduleDiff, status := calculateScheduleStatus(profile.Challenges[i])
			profile.Challenges[i].Progress.Rate.ScheduleDiff = scheduleDiff
			profile.Challenges[i].Progress.Rate.Status = status

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

// updateChallenges recalculates progress for each reading challenge in the profile.
// It uses your existing calculation functions (calculateCurrentPace and calculateScheduleStatus)
// to update the challenge fields.
func updateChallenges(profile *models.Profile) {
	now := time.Now()
	// Loop through every challenge on the profile
	for i, ch := range profile.Challenges {
		log.Printf("Updating challenge %s: target=%d, type=%s, timeframe=%s", ch.ID, ch.Target, ch.Type, ch.TimeFrame)

		// Compute the aggregated progress based on challenge type.
		aggProgress := aggregateChallengeProgress(profile, ch)
		log.Printf("Aggregated progress for challenge %s: %d", ch.ID, aggProgress)

		// Set the current progress and update percentage.
		profile.Challenges[i].Progress.Current = aggProgress
		if ch.Target != 0 {
			profile.Challenges[i].Progress.Percentage = float64(aggProgress) / float64(ch.Target) * 100
		}

		// Update the reading pace. (calculateCurrentPace should use the challenge's start date and current progress.)
		currentPace := calculateCurrentPace(ch)
		profile.Challenges[i].Progress.Rate.CurrentPace = currentPace

		// Calculate the schedule difference and status.
		scheduleDiff, status := calculateScheduleStatus(ch)
		profile.Challenges[i].Progress.Rate.ScheduleDiff = scheduleDiff
		profile.Challenges[i].Progress.Rate.Status = status

		// Update the challenge's timestamp.
		profile.Challenges[i].UpdatedAt = now
		log.Printf("Updated challenge %s: current=%d, percentage=%.2f%%, current pace=%.2f, schedule diff=%.2f, status=%s",
			ch.ID, aggProgress, profile.Challenges[i].Progress.Percentage, currentPace, scheduleDiff, status)
	}
}

// aggregateChallengeProgress aggregates the total progress for a given challenge.
// It differentiates based on the challenge type (books vs pages) and logs the result.
// For a books challenge, it counts the number of reading log entries with a note of "Book Finished".
// For a pages challenge, it sums the PagesRead values.
func aggregateChallengeProgress(profile *models.Profile, challenge models.ReadingChallenge) int {
	total := 0
	switch challenge.Type {
	case models.BooksChallenge:
		// For a books challenge, count the number of log entries that indicate a book was completed.
		for _, logEntry := range profile.ReadingLog {
			// Adjust the logic as needed to match your book‐completion criteria.
			if logEntry.Notes == "Book Finished" {
				total++
			}
		}
		log.Printf("Challenge %s (Books): total completed books = %d", challenge.ID, total)
	case models.PagesChallenge:
		// For a pages challenge, sum the pages read.
		for _, logEntry := range profile.ReadingLog {
			total += logEntry.PagesRead
		}
		log.Printf("Challenge %s (Pages): total pages read = %d", challenge.ID, total)
	default:
		// If you have other types or want a default behavior, you could add it here.
		log.Printf("Challenge %s: unknown type %s; defaulting aggregated progress to 0", challenge.ID, challenge.Type)
	}
	return total
}
