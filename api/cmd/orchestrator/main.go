// cmd/bookit-lambda/main.go
package main

import (
	"log"
	"os"

	"github.com/FriedGlue/BookIt/api/pkg/delivery"
	"github.com/FriedGlue/BookIt/api/pkg/repository/dynamodb"
	"github.com/FriedGlue/BookIt/api/pkg/service"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws/session"
)

func main() {
	log.Println("Starting Orchestrator Lambda function")

	sess := session.Must(session.NewSession())
	log.Println("AWS Session created successfully")

	// Check for environment variables
	booksTable := os.Getenv("BOOKS_TABLE_NAME")
	profilesTable := os.Getenv("PROFILES_TABLE_NAME")

	if booksTable == "" {
		log.Println("BOOKS_TABLE_NAME environment variable not set, using BOOKS_TABLE as fallback")
		booksTable = os.Getenv("BOOKS_TABLE")
	}

	if profilesTable == "" {
		log.Println("PROFILES_TABLE_NAME environment variable not set, using PROFILES_TABLE as fallback")
		profilesTable = os.Getenv("PROFILES_TABLE")
	}

	log.Printf("Using Books Table: %s", booksTable)
	log.Printf("Using Profiles Table: %s", profilesTable)

	// Initialize repositories
	bookRepo := dynamodb.NewBookRepo(sess, booksTable)
	profileRepo := dynamodb.NewProfileRepo(sess, profilesTable)
	log.Println("Repositories initialized")

	// Initialize services - each service is created exactly once
	logSvc := service.NewReadingLogService(profileRepo)
	bookSvc := service.NewBookService(bookRepo)
	profileSvc := service.NewProfileService(profileRepo)
	currentlyReadingSvc := service.NewCurrentlyReadingService(profileRepo, logSvc, bookRepo)
	readingListSvc := service.NewReadingListService(profileRepo, bookRepo)
	challengeSvc := service.NewChallengeService(profileRepo, readingListSvc)

	log.Println("Services initialized")

	// Initialize router with all services
	handler := delivery.NewRouter(
		bookSvc,
		profileSvc,
		logSvc,
		currentlyReadingSvc,
		readingListSvc,
		challengeSvc,
	)
	log.Println("Router initialized, starting Lambda handler")

	lambda.Start(handler.Handler)
}
