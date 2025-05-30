package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"math/rand"
	"time"

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
		return shared.ErrorResponse(401, err.Error())
	}

	svc := shared.DynamoDBClient()
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
		return shared.ErrorResponse(500, fmt.Sprintf("DynamoDB GetItem error: %v", err))
	}
	if result.Item == nil {
		log.Println("Profile not found")
		return shared.ErrorResponse(404, "Profile not found")
	}

	var profile models.Profile
	if err := dynamodbattribute.UnmarshalMap(result.Item, &profile); err != nil {
		log.Printf("Error unmarshalling profile: %v\n", err)
		return shared.ErrorResponse(500, "Error unmarshalling profile: "+err.Error())
	}

	responseBody, err := json.Marshal(profile.CurrentlyReading)
	if err != nil {
		log.Printf("Error marshalling response: %v\n", err)
		return shared.ErrorResponse(500, "Error marshalling currently reading response")
	}

	log.Println("Currently reading list retrieval successful")
	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Body:       string(responseBody),
	}
}

type newCurrentlyReadingItemRequest struct {
	ISBN   string `json:"isbn"`
	BookID string `json:"bookId,omitempty"`
}

// AddToCurrentlyReading adds a new currentlyReadingItem to the "currently reading" list in the Profile table
func AddToCurrentlyReading(request events.APIGatewayProxyRequest) events.APIGatewayProxyResponse {
	log.Println("AddToCurrentlyReading invoked")
	userId, err := shared.GetUserIDFromToken(request)
	if err != nil {
		log.Printf("Error extracting userId: %v\n", err)
		return shared.ErrorResponse(401, err.Error())
	}

	var newCurrentlyReadingItemRequest newCurrentlyReadingItemRequest
	if err := json.Unmarshal([]byte(request.Body), &newCurrentlyReadingItemRequest); err != nil {
		log.Printf("Invalid JSON: %v\n", err)
		return shared.ErrorResponse(400, "Invalid JSON: "+err.Error())
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
	if result.Item == nil {
		log.Println("Profile not found")
		return shared.ErrorResponse(404, "Profile not found")
	}

	var profile models.Profile
	if err := dynamodbattribute.UnmarshalMap(result.Item, &profile); err != nil {
		log.Printf("Error unmarshalling profile: %v\n", err)
		return shared.ErrorResponse(500, "Error unmarshalling profile: "+err.Error())
	}

	found := false
	for _, item := range profile.CurrentlyReading {
		if item.Book.ISBN == newCurrentlyReadingItemRequest.ISBN || item.Book.BookID == newCurrentlyReadingItemRequest.BookID {
			found = true
			log.Printf("Book (ID: %s, ISBN: %s) already exists in user's currently reading list",
				newCurrentlyReadingItemRequest.BookID, newCurrentlyReadingItemRequest.ISBN)
			break
		}
	}

	if found {
		return shared.ErrorResponse(409, "Book already in currently reading list")
	}

	var bookDetails BookData
	if newCurrentlyReadingItemRequest.BookID != "" {
		// If the book ID is provided, we need to fetch the book details from the Books table
		getBookInput := &dynamodb.GetItemInput{
			TableName: aws.String(BOOKS_TABLE_NAME),
			Key: map[string]*dynamodb.AttributeValue{
				"bookId": {S: aws.String(newCurrentlyReadingItemRequest.BookID)},
			},
		}
		bookResult, err := svc.GetItem(getBookInput)
		if err != nil {
			log.Printf("DynamoDB GetItem error: %v\n", err)
			return shared.ErrorResponse(500, fmt.Sprintf("DynamoDB GetItem error: %v", err))
		}
		if bookResult.Item == nil {
			log.Println("Book not found")
			return shared.ErrorResponse(404, "Book not found")
		}
		if err := dynamodbattribute.UnmarshalMap(bookResult.Item, &bookDetails); err != nil {
			log.Printf("Error unmarshalling book details: %v\n", err)
			return shared.ErrorResponse(500, "Error unmarshalling book details: "+err.Error())
		}
	} else {
		// If the book ID is not provided, we need to fetch the book details from the openlibrary API
		bookDetails, err = FetchBookFromOpenLibrary(newCurrentlyReadingItemRequest.ISBN)
		if err != nil {
			log.Printf("Error fetching book details: %v\n", err)
			return shared.ErrorResponse(500, "Error fetching book details: "+err.Error())
		}
	}

	// Create a new CurrentlyReadingItem and add it to the profile using the book details
	temp := rand.New(rand.NewSource(time.Now().UnixNano()))
	book := models.Book{
		BookID:     fmt.Sprintf("%d", temp.Int()),
		ISBN:       bookDetails.ISBN13,
		Title:      bookDetails.Title,
		Authors:    bookDetails.Authors,
		Thumbnail:  bookDetails.CoverImageURL,
		TotalPages: bookDetails.PageCount,
		Progress: models.ReadingProgress{
			LastPageRead: 0,
			Percentage:   0,
			LastUpdated:  time.Now().Format(time.RFC3339),
		},
	}
	currentlyReadingItem := models.CurrentlyReadingItem{
		Book:        book,
		StartedDate: time.Now().Format(time.RFC3339),
	}

	// Set a default page count if it's zero
	if currentlyReadingItem.Book.TotalPages == 0 {
		log.Printf("TotalPages for book is 0, setting default value of 300\n")
		currentlyReadingItem.Book.TotalPages = 300
	}

	profile.CurrentlyReading = append(profile.CurrentlyReading, currentlyReadingItem)

	// Update the reading log with the new progress
	logEntry := models.ReadingLogItem{
		Id:            fmt.Sprintf("%d", rand.Int()),
		Date:          time.Now().Format(time.RFC3339),
		BookID:        book.BookID,
		Title:         book.Title,
		BookThumbnail: book.Thumbnail,
		PagesRead:     0,
		Notes:         "Book Started",
	}
	profile.ReadingLog = append(profile.ReadingLog, logEntry)
	updateChallenges(&profile)

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

	log.Printf("Book added to currently reading for user %s\n", userId)
	return events.APIGatewayProxyResponse{
		StatusCode: 201,
		Body:       "Book added to currently reading",
	}
}

// UpdateCurrentlyReading updates a book in the "currently reading" list in the Profile table
type updateCurrentlyReadingRequest struct {
	ISBN        string `json:"isbn,omitempty"`
	CurrentPage int    `json:"currentPage"`
	BookID      string `json:"bookId,omitempty"`
	Title       string `json:"title,omitempty"`
	Notes       string `json:"notes,omitempty"`
}

// UpdateCurrentlyReading updates a book in the "currently reading" list in the Profile table
func UpdateCurrentlyReading(request events.APIGatewayProxyRequest) events.APIGatewayProxyResponse {
	log.Println("UpdateCurrentlyReading invoked")

	// Extract userId from token
	userId, err := shared.GetUserIDFromToken(request)
	if err != nil {
		log.Printf("Error extracting userId: %v\n", err)
		return shared.ErrorResponse(401, err.Error())
	}
	log.Printf("Extracted userId: %s\n", userId)

	// Parse request body to updateCurrentlyReadingRequest struct
	var updateReq updateCurrentlyReadingRequest
	if err := json.Unmarshal([]byte(request.Body), &updateReq); err != nil {
		log.Printf("Invalid JSON in request body: %v\n", err)
		return shared.ErrorResponse(400, "Invalid JSON: "+err.Error())
	}
	log.Printf("Parsed update request: %+v\n", updateReq)

	svc := shared.DynamoDBClient()

	// Get the current profile from DynamoDB
	getInput := &dynamodb.GetItemInput{
		TableName: aws.String(PROFILES_TABLE_NAME),
		Key: map[string]*dynamodb.AttributeValue{
			"_id": {S: aws.String(userId)},
		},
	}
	log.Printf("Attempting to retrieve profile for userId: %s\n", userId)
	result, err := svc.GetItem(getInput)
	if err != nil {
		log.Printf("DynamoDB GetItem error: %v\n", err)
		return shared.ErrorResponse(500, fmt.Sprintf("DynamoDB GetItem error: %v", err))
	}
	if result.Item == nil {
		log.Println("Profile not found")
		return shared.ErrorResponse(404, "Profile not found")
	}
	log.Printf("Profile retrieved: %v\n", result.Item)

	// Unmarshal DynamoDB result into profile struct
	var profile models.Profile
	if err := dynamodbattribute.UnmarshalMap(result.Item, &profile); err != nil {
		log.Printf("Error unmarshalling profile: %v\n", err)
		return shared.ErrorResponse(500, "Error unmarshalling profile: "+err.Error())
	}
	log.Printf("Unmarshalled profile: %+v\n", profile)

	// Find the book in the currently reading list
	found := false
	bookIndex := -1

	// Log currently reading books to help debug
	log.Printf("Current currently reading books count: %d", len(profile.CurrentlyReading))
	for i, item := range profile.CurrentlyReading {
		log.Printf("Book %d: ID=%s, ISBN=%s, Title=%s", i, item.Book.BookID, item.Book.ISBN, item.Book.Title)
	}
	log.Printf("Looking for book with BookID=%s or ISBN=%s", updateReq.BookID, updateReq.ISBN)

	for i, item := range profile.CurrentlyReading {
		// Log each comparison to help debug
		if updateReq.BookID != "" {
			log.Printf("Comparing book %d: provided bookId '%s' == item.Book.BookID '%s': %v",
				i, updateReq.BookID, item.Book.BookID, item.Book.BookID == updateReq.BookID)
		}
		if updateReq.ISBN != "" {
			log.Printf("Comparing book %d: provided ISBN '%s' == item.Book.ISBN '%s': %v",
				i, updateReq.ISBN, item.Book.ISBN, item.Book.ISBN == updateReq.ISBN)
		}

		// Match by BookID or ISBN
		matchesBookId := updateReq.BookID != "" && item.Book.BookID == updateReq.BookID
		matchesISBN := updateReq.ISBN != "" && item.Book.ISBN == updateReq.ISBN

		if matchesBookId || matchesISBN {
			bookIndex = i
			found = true
			log.Printf("Found matching book at index %d: BookID=%s, ISBN=%s, Title=%s",
				i, item.Book.BookID, item.Book.ISBN, item.Book.Title)
			break
		}
	}

	if !found {
		log.Printf("Book with ISBN %s or BookID %s not found in currently reading list\n", updateReq.ISBN, updateReq.BookID)
		return shared.ErrorResponse(404, "Book not found in currently reading list")
	}

	// Calculate new progress percentage
	if profile.CurrentlyReading[bookIndex].Book.TotalPages == 0 {
		log.Printf("TotalPages for book is 0, setting default value of 300\n")
		// Set a default page count instead of failing
		profile.CurrentlyReading[bookIndex].Book.TotalPages = 300
	}

	newProgressPercentage := math.Floor(
		float64(updateReq.CurrentPage) / float64(profile.CurrentlyReading[bookIndex].Book.TotalPages) * 100,
	)
	log.Printf("Calculated new progress percentage: %.2f%% for currentPage: %d and totalPages: %d\n",
		newProgressPercentage, updateReq.CurrentPage, profile.CurrentlyReading[bookIndex].Book.TotalPages)

	// Calculate the number of pages read
	pagesRead := updateReq.CurrentPage - profile.CurrentlyReading[bookIndex].Book.Progress.LastPageRead

	// Update the progress data directly in the array
	profile.CurrentlyReading[bookIndex].Book.Progress.LastPageRead = updateReq.CurrentPage
	profile.CurrentlyReading[bookIndex].Book.Progress.Percentage = newProgressPercentage
	profile.CurrentlyReading[bookIndex].Book.Progress.LastUpdated = time.Now().Format(time.RFC3339)
	log.Printf("Updated book progress: %+v\n", profile.CurrentlyReading[bookIndex].Book.Progress)

	// Update the reading log with the new progress
	logEntry := models.ReadingLogItem{
		Id:            fmt.Sprintf("%d", rand.Int()),
		Date:          time.Now().Format(time.RFC3339),
		BookID:        profile.CurrentlyReading[bookIndex].Book.BookID,
		Title:         profile.CurrentlyReading[bookIndex].Book.Title,
		BookThumbnail: profile.CurrentlyReading[bookIndex].Book.Thumbnail,
		PagesRead:     pagesRead,
		Notes:         updateReq.Notes,
	}
	profile.ReadingLog = append(profile.ReadingLog, logEntry)
	updateChallenges(&profile)

	// Marshal the updated profile back to DynamoDB format
	updatedProfile, err := dynamodbattribute.MarshalMap(profile)
	if err != nil {
		log.Printf("Error marshalling updated profile: %v\n", err)
		return shared.ErrorResponse(500, "Error marshalling updated profile: "+err.Error())
	}
	log.Printf("Marshalled updated profile: %v\n", updatedProfile)

	// Put the updated profile back into DynamoDB
	putInput := &dynamodb.PutItemInput{
		TableName: aws.String(PROFILES_TABLE_NAME),
		Item:      updatedProfile,
	}
	_, err = svc.PutItem(putInput)
	if err != nil {
		log.Printf("DynamoDB PutItem error: %v\n", err)
		return shared.ErrorResponse(500, fmt.Sprintf("DynamoDB PutItem error: %v", err))
	}
	log.Printf("Successfully updated book progress in DynamoDB for user %s\n", userId)

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
		return shared.ErrorResponse(401, err.Error())
	}

	bookId := request.QueryStringParameters["bookId"]
	if bookId == "" {
		return shared.ErrorResponse(400, "bookId query parameter is required")
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
	if result.Item == nil {
		log.Println("Profile not found")
		return shared.ErrorResponse(404, "Profile not found")
	}

	var profile models.Profile
	if err := dynamodbattribute.UnmarshalMap(result.Item, &profile); err != nil {
		log.Printf("Error unmarshalling profile: %v\n", err)
		return shared.ErrorResponse(500, "Error unmarshalling profile: "+err.Error())
	}

	// Log the current state of currently reading books to help debug
	log.Printf("Current currently reading books count: %d", len(profile.CurrentlyReading))
	for i, item := range profile.CurrentlyReading {
		log.Printf("Book %d: ID=%s, Title=%s", i, item.Book.BookID, item.Book.Title)
	}
	log.Printf("Looking for book with ID=%s", bookId)

	var bookDetails models.Book
	index := -1
	for i, item := range profile.CurrentlyReading {
		// Log each comparison to debug the matching logic
		log.Printf("Comparing book %d: provided bookId '%s' == item.Book.BookID '%s': %v",
			i, bookId, item.Book.BookID, item.Book.BookID == bookId)

		if item.Book.BookID == bookId {
			index = i
			bookDetails = item.Book
			log.Printf("Found match at index %d: BookID=%s, Title=%s",
				i, item.Book.BookID, item.Book.Title)
			break
		}
	}

	if index == -1 {
		log.Printf("Book not found in currently reading list for user %s\n", userId)
		return shared.ErrorResponse(404, "Book not found in currently reading list")
	}

	// Log before removal
	log.Printf("Removing book at index %d from currently reading list", index)
	profile.CurrentlyReading = append(profile.CurrentlyReading[:index], profile.CurrentlyReading[index+1:]...)
	// Update the reading log with the new progress
	logEntry := models.ReadingLogItem{
		Id:            fmt.Sprintf("%d", rand.Int()),
		Date:          time.Now().Format(time.RFC3339),
		BookID:        bookDetails.BookID,
		Title:         bookDetails.Title,
		BookThumbnail: bookDetails.Thumbnail,
		PagesRead:     bookDetails.Progress.LastPageRead,
		Notes:         "Book Removed",
	}
	profile.ReadingLog = append(profile.ReadingLog, logEntry)
	updateChallenges(&profile)

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

	log.Printf("Book removed from currently reading for user %s\n", userId)
	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Body:       "Book removed from currently reading",
	}
}

// StartReadingRequest represents the request body for starting a book
type StartReadingRequest struct {
	BookID   string `json:"bookId"`
	ListName string `json:"listName"` // "toBeRead", "read", or custom list name
}

// StartReading moves a book from any list to currently reading
func StartReading(request events.APIGatewayProxyRequest) events.APIGatewayProxyResponse {
	log.Println("StartReading invoked")
	userId, err := shared.GetUserIDFromToken(request)
	if err != nil {
		log.Printf("Error extracting userId: %v\n", err)
		return shared.ErrorResponse(401, err.Error())
	}

	var startReq StartReadingRequest
	if err := json.Unmarshal([]byte(request.Body), &startReq); err != nil {
		log.Printf("Invalid JSON: %v\n", err)
		return shared.ErrorResponse(400, "Invalid JSON: "+err.Error())
	}

	if startReq.BookID == "" {
		return shared.ErrorResponse(400, "bookId is required")
	}
	if startReq.ListName == "" {
		return shared.ErrorResponse(400, "listName is required")
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
	if result.Item == nil {
		log.Println("Profile not found")
		return shared.ErrorResponse(404, "Profile not found")
	}

	var profile models.Profile
	if err := dynamodbattribute.UnmarshalMap(result.Item, &profile); err != nil {
		log.Printf("Error unmarshalling profile: %v\n", err)
		return shared.ErrorResponse(500, "Error unmarshalling profile: "+err.Error())
	}

	// Find and remove the book from the specified list
	found := false

	// Special case for "direct" list name - this means add directly without checking any list
	if startReq.ListName == "direct" {
		// When coming directly from book detail page, we don't need to look for the book in a list
		found = true
		log.Printf("Using direct mode - skipping list check")
	} else {
		// Normal flow - look for and remove from the specified list
		switch startReq.ListName {
		case "toBeRead":
			for i, item := range profile.Lists.ToBeRead {
				if item.BookID == startReq.BookID {
					profile.Lists.ToBeRead = append(profile.Lists.ToBeRead[:i], profile.Lists.ToBeRead[i+1:]...)
					found = true
					break
				}
			}
		case "read":
			for i, item := range profile.Lists.Read {
				if item.BookID == startReq.BookID {
					profile.Lists.Read = append(profile.Lists.Read[:i], profile.Lists.Read[i+1:]...)
					found = true
					break
				}
			}
		default:
			// Check custom lists
			if customList, exists := profile.Lists.CustomLists[startReq.ListName]; exists {
				for i, item := range customList {
					if item.BookID == startReq.BookID {
						profile.Lists.CustomLists[startReq.ListName] = append(customList[:i], customList[i+1:]...)
						found = true
						break
					}
				}
			}
		}
	}

	if !found {
		return shared.ErrorResponse(404, fmt.Sprintf("Book not found in %s list", startReq.ListName))
	}

	// Get book details from the Books table
	getBookInput := &dynamodb.GetItemInput{
		TableName: aws.String(BOOKS_TABLE_NAME),
		Key: map[string]*dynamodb.AttributeValue{
			"bookId": {S: aws.String(startReq.BookID)},
		},
	}
	bookResult, err := svc.GetItem(getBookInput)
	if err != nil {
		log.Printf("DynamoDB GetItem error: %v\n", err)
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

	// Create a new currently reading item
	currentlyReadingItem := models.CurrentlyReadingItem{
		Book: models.Book{
			BookID:     bookDetails.BookID,
			ISBN:       bookDetails.ISBN13,
			Title:      bookDetails.Title,
			Authors:    bookDetails.Authors,
			Thumbnail:  bookDetails.CoverImageURL,
			TotalPages: bookDetails.PageCount,
			Progress: models.ReadingProgress{
				LastPageRead: 0,
				Percentage:   0,
				LastUpdated:  time.Now().Format(time.RFC3339),
			},
		},
		StartedDate: time.Now().Format(time.RFC3339),
	}

	// Set a default page count if it's zero
	if currentlyReadingItem.Book.TotalPages == 0 {
		log.Printf("TotalPages for book is 0, setting default value of 300\n")
		currentlyReadingItem.Book.TotalPages = 300
	}

	// Add to currently reading list
	profile.CurrentlyReading = append(profile.CurrentlyReading, currentlyReadingItem)

	// Update the reading log with the new progress
	logEntry := models.ReadingLogItem{
		Id:            fmt.Sprintf("%d", rand.Int()),
		Date:          time.Now().Format(time.RFC3339),
		BookID:        bookDetails.BookID,
		Title:         bookDetails.Title,
		BookThumbnail: bookDetails.CoverImageURL,
		PagesRead:     0,
		Notes:         "Book Started",
	}
	profile.ReadingLog = append(profile.ReadingLog, logEntry)
	updateChallenges(&profile)
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

	// Different message based on whether we moved from a list or added directly
	if startReq.ListName == "direct" {
		log.Printf("Book added directly to currently reading for user %s\n", userId)
		return events.APIGatewayProxyResponse{
			StatusCode: 200,
			Body:       "Book added directly to currently reading list",
		}
	} else if found {
		log.Printf("Book moved to currently reading from %s list for user %s\n", startReq.ListName, userId)
		return events.APIGatewayProxyResponse{
			StatusCode: 200,
			Body:       fmt.Sprintf("Book moved to currently reading from %s list", startReq.ListName),
		}
	} else {
		log.Printf("Book added to currently reading for user %s\n", userId)
		return events.APIGatewayProxyResponse{
			StatusCode: 200,
			Body:       "Book added to currently reading list",
		}
	}
}

// FinishReadingRequest represents the request body for finishing a book
type FinishReadingRequest struct {
	BookID string `json:"bookId"`
}

// FinishReading moves a book from currently reading to read list
func FinishReading(request events.APIGatewayProxyRequest) events.APIGatewayProxyResponse {
	log.Println("FinishReading invoked")
	userId, err := shared.GetUserIDFromToken(request)
	if err != nil {
		log.Printf("Error extracting userId: %v\n", err)
		return shared.ErrorResponse(401, err.Error())
	}

	var finishReq FinishReadingRequest
	if err := json.Unmarshal([]byte(request.Body), &finishReq); err != nil {
		log.Printf("Invalid JSON: %v\n", err)
		return shared.ErrorResponse(400, "Invalid JSON: "+err.Error())
	}

	if finishReq.BookID == "" {
		return shared.ErrorResponse(400, "bookId is required")
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
	if result.Item == nil {
		log.Println("Profile not found")
		return shared.ErrorResponse(404, "Profile not found")
	}

	var profile models.Profile
	if err := dynamodbattribute.UnmarshalMap(result.Item, &profile); err != nil {
		log.Printf("Error unmarshalling profile: %v\n", err)
		return shared.ErrorResponse(500, "Error unmarshalling profile: "+err.Error())
	}

	// Find and remove the book from currently reading
	var bookToMove models.CurrentlyReadingItem
	found := false
	for i, item := range profile.CurrentlyReading {
		if item.Book.BookID == finishReq.BookID {
			bookToMove = item
			profile.CurrentlyReading = append(profile.CurrentlyReading[:i], profile.CurrentlyReading[i+1:]...)
			found = true
			break
		}
	}

	if !found {
		return shared.ErrorResponse(404, "Book not found in currently reading list")
	}

	// Create a new read item
	readItem := models.ReadItem{
		BookID:        bookToMove.Book.BookID,
		CompletedDate: time.Now().Format(time.RFC3339),
		Title:         bookToMove.Book.Title,
		Authors:       bookToMove.Book.Authors,
		Thumbnail:     bookToMove.Book.Thumbnail,
		Rating:        0,  // Initial rating
		Review:        "", // Initial review
		Order:         len(profile.Lists.Read),
	}

	// Initialize Lists if needed and add to read list
	if len(profile.Lists.Read) == 0 {
		profile.Lists.Read = []models.ReadItem{}
	}
	profile.Lists.Read = append(profile.Lists.Read, readItem)

	// Update the reading log with the new progress
	logEntry := models.ReadingLogItem{
		Id:            fmt.Sprintf("%d", rand.Int()),
		Date:          time.Now().Format(time.RFC3339),
		BookID:        readItem.BookID,
		Title:         readItem.Title,
		BookThumbnail: readItem.Thumbnail,
		PagesRead:     bookToMove.Book.TotalPages,
		Notes:         "Book Finished",
	}
	profile.ReadingLog = append(profile.ReadingLog, logEntry)
	updateChallenges(&profile)
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

	log.Printf("Book moved to read list for user %s\n", userId)
	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Body:       "Book moved to read list",
	}
}
