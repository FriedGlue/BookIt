package main

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/dynamodb/expression"
)

type UserList struct {
	UserID      string   `json:"userId"`          // Partition Key
	ListName    string   `json:"listName"`        // Sort Key
	ISBNs       []string `json:"isbns,omitempty"` // Could be stored as a set (if so, these are string slices)
	LastUpdated string   `json:"lastUpdated,omitempty"`
	Description string   `json:"description,omitempty"`
}

func getUserIDFromToken(request events.APIGatewayProxyRequest) (string, error) {
	// The "sub" claim in Cognito is typically the unique user identifier
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

// ---------------- GET (Read) ----------------

func GetUserLists(request events.APIGatewayProxyRequest) events.APIGatewayProxyResponse {
	userId, err := getUserIDFromToken(request)
	if err != nil {
		return errorResponse(401, err.Error())
	}

	listName := request.PathParameters["listName"]

	svc := DynamoDBClient()

	// If listName is provided => single item
	if listName != "" {
		getInput := &dynamodb.GetItemInput{
			TableName: aws.String("UserLists"),
			Key: map[string]*dynamodb.AttributeValue{
				"userId":   {S: aws.String(userId)},
				"listName": {S: aws.String(listName)},
			},
		}
		result, err := svc.GetItem(getInput)
		if err != nil {
			return errorResponse(500, "DynamoDB GetItem error: "+err.Error())
		}
		if result.Item == nil {
			return errorResponse(404, "List not found")
		}

		var userList UserList
		if err := dynamodbattribute.UnmarshalMap(result.Item, &userList); err != nil {
			return errorResponse(500, "Unmarshal error: "+err.Error())
		}

		// Optionally fetch book metadata with BatchGetItem if needed
		// ...

		bytes, _ := json.Marshal(userList)
		return events.APIGatewayProxyResponse{
			StatusCode: 200,
			Body:       string(bytes),
		}
	}

	// Otherwise => Query all lists for the user
	queryInput := &dynamodb.QueryInput{
		TableName:              aws.String("UserLists"),
		KeyConditionExpression: aws.String("userId = :uid"),
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":uid": {S: aws.String(userId)},
		},
	}
	result, err := svc.Query(queryInput)
	if err != nil {
		return errorResponse(500, "Query error: "+err.Error())
	}
	if len(result.Items) == 0 {
		return errorResponse(404, "No lists found for user")
	}

	var lists []UserList
	if err := dynamodbattribute.UnmarshalListOfMaps(result.Items, &lists); err != nil {
		return errorResponse(500, "Unmarshal error: "+err.Error())
	}

	bytes, _ := json.Marshal(lists)
	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Body:       string(bytes),
	}
}

// ---------------- POST (Create) ----------------

func CreateUserList(request events.APIGatewayProxyRequest) events.APIGatewayProxyResponse {
	userId, err := getUserIDFromToken(request)
	if err != nil {
		return errorResponse(401, err.Error())
	}

	listName, ok := request.PathParameters["listName"]
	if !ok || listName == "" {
		return errorResponse(400, "Missing listName in path")
	}

	var payload struct {
		ISBNs       []string `json:"isbns"`
		Description string   `json:"description"`
	}
	if err := json.Unmarshal([]byte(request.Body), &payload); err != nil {
		return errorResponse(400, "Invalid JSON: "+err.Error())
	}

	userList := UserList{
		UserID:      userId,
		ListName:    listName,
		ISBNs:       payload.ISBNs,
		Description: payload.Description,
		LastUpdated: time.Now().Format(time.RFC3339),
	}

	av, err := dynamodbattribute.MarshalMap(userList)
	if err != nil {
		return errorResponse(500, "Marshal error: "+err.Error())
	}

	svc := DynamoDBClient()
	input := &dynamodb.PutItemInput{
		TableName: aws.String("UserLists"),
		Item:      av,
	}

	if _, err := svc.PutItem(input); err != nil {
		return errorResponse(500, "DynamoDB PutItem error: "+err.Error())
	}

	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Body:       fmt.Sprintf("List '%s' created for user %s", listName, userId),
	}
}

// ---------------- PUT (Update) ----------------

func UpdateUserList(request events.APIGatewayProxyRequest) events.APIGatewayProxyResponse {
	userId, err := getUserIDFromToken(request)
	if err != nil {
		return errorResponse(401, err.Error())
	}

	listName, ok := request.PathParameters["listName"]
	if !ok || listName == "" {
		return errorResponse(400, "Missing listName in path")
	}

	// The payload can contain optional fields for partial updates
	var payload struct {
		AddISBNs    []string `json:"addIsbns,omitempty"`    // ISBNs to add
		RemoveISBNs []string `json:"removeIsbns,omitempty"` // ISBNs to remove
		Description *string  `json:"description,omitempty"` // optional new description
	}
	if err := json.Unmarshal([]byte(request.Body), &payload); err != nil {
		return errorResponse(400, "Invalid JSON: "+err.Error())
	}

	updateBuilder := expression.UpdateBuilder{}

	// If you stored 'isbns' as a DynamoDB *string set*, you can use ADD/DELETE operators:
	if len(payload.AddISBNs) > 0 {
		updateBuilder = updateBuilder.Add(
			expression.Name("isbns"),
			expression.Value(&dynamodb.AttributeValue{SS: aws.StringSlice(payload.AddISBNs)}),
		)
	}
	if len(payload.RemoveISBNs) > 0 {
		updateBuilder = updateBuilder.Delete(
			expression.Name("isbns"),
			expression.Value(&dynamodb.AttributeValue{SS: aws.StringSlice(payload.RemoveISBNs)}),
		)
	}
	// Alternatively, if 'isbns' is a list, you might have to read, modify, and rewrite the entire list.

	// Update description if provided
	if payload.Description != nil {
		updateBuilder = updateBuilder.Set(
			expression.Name("description"),
			expression.Value(*payload.Description),
		)
	}

	// Always update lastUpdated timestamp
	updateBuilder = updateBuilder.Set(
		expression.Name("lastUpdated"),
		expression.Value(time.Now().Format(time.RFC3339)),
	)

	expr, err := expression.NewBuilder().WithUpdate(updateBuilder).Build()
	if err != nil {
		return errorResponse(500, "Error building update expression: "+err.Error())
	}

	svc := DynamoDBClient()
	input := &dynamodb.UpdateItemInput{
		TableName: aws.String("UserLists"),
		Key: map[string]*dynamodb.AttributeValue{
			"userId":   {S: aws.String(userId)},
			"listName": {S: aws.String(listName)},
		},
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		UpdateExpression:          expr.Update(),
		ReturnValues:              aws.String("UPDATED_NEW"),
	}

	if _, err := svc.UpdateItem(input); err != nil {
		return errorResponse(500, "Error updating list: "+err.Error())
	}

	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Body:       fmt.Sprintf("List '%s' updated for user %s", listName, userId),
	}
}

// ---------------- DELETE (Remove) ----------------

func DeleteUserList(request events.APIGatewayProxyRequest) events.APIGatewayProxyResponse {
	userId, err := getUserIDFromToken(request)
	if err != nil {
		return errorResponse(401, err.Error())
	}

	listName, ok := request.PathParameters["listName"]
	if !ok || listName == "" {
		return errorResponse(400, "Missing listName in path")
	}

	svc := DynamoDBClient()
	input := &dynamodb.DeleteItemInput{
		TableName: aws.String("UserLists"),
		Key: map[string]*dynamodb.AttributeValue{
			"userId":   {S: aws.String(userId)},
			"listName": {S: aws.String(listName)},
		},
	}

	if _, err := svc.DeleteItem(input); err != nil {
		return errorResponse(500, "Error deleting list: "+err.Error())
	}

	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Body:       fmt.Sprintf("List '%s' deleted for user %s", listName, userId),
	}
}
