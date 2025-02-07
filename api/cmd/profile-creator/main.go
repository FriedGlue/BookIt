package main

import (
	"context"
	"encoding/json"
	"log"
	"os"

	"github.com/FriedGlue/BookIt/api/pkg/models"
	"github.com/FriedGlue/BookIt/api/pkg/shared"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

type UserEvent struct {
	Action   string `json:"action"`
	Username string `json:"username"`
	Sub      string `json:"sub"`
}

func handleRequest(ctx context.Context, snsEvent events.SNSEvent) error {
	log.Printf("Processing %d records\n", len(snsEvent.Records))

	for _, record := range snsEvent.Records {
		var userEvent UserEvent
		if err := json.Unmarshal([]byte(record.SNS.Message), &userEvent); err != nil {
			log.Printf("Error unmarshalling message: %v", err)
			continue
		}

		if userEvent.Action != "CREATE_USER_PROFILE" {
			log.Printf("Skipping unknown action: %s", userEvent.Action)
			continue
		}

		// Create a new profile
		profile := models.Profile{
			ID: userEvent.Sub,
			ProfileInformation: models.ProfileInformation{
				Username: userEvent.Username,
			},
			CurrentlyReading: []models.CurrentlyReadingItem{},
			Lists: models.UserLists{
				ToBeRead:    []models.ToBeReadItem{},
				Read:        []models.ReadItem{},
				CustomLists: make(map[string][]models.CustomListItem),
			},
			ReadingLog: []models.ReadingLogItem{},
			Challenges: []models.ReadingChallenge{},
		}

		// Marshal the profile to DynamoDB format
		item, err := dynamodbattribute.MarshalMap(profile)
		if err != nil {
			log.Printf("Error marshalling profile: %v", err)
			continue
		}

		// Save to DynamoDB
		svc := shared.DynamoDBClient()
		input := &dynamodb.PutItemInput{
			TableName: aws.String(os.Getenv("PROFILES_TABLE_NAME")),
			Item:      item,
		}

		if _, err := svc.PutItem(input); err != nil {
			log.Printf("Error saving profile: %v", err)
			continue
		}

		log.Printf("Successfully created profile for user: %s", userEvent.Username)
	}

	return nil
}

func main() {
	lambda.Start(handleRequest)
}
