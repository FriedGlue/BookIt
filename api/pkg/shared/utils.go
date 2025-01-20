package shared

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
)

// getUserIDFromToken extracts the user’s Cognito “sub” claim
func GetUserIDFromToken(request events.APIGatewayProxyRequest) (string, error) {
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

// DynamoDBClient initializes a DynamoDB client session.
func DynamoDBClient() *dynamodb.DynamoDB {
	sess := session.Must(session.NewSession())
	return dynamodb.New(sess)
}

// ErrorResponse is a helper to generate an APIGatewayProxyResponse with a given status and message.
func ErrorResponse(status int, message string) events.APIGatewayProxyResponse {
	body, _ := json.Marshal(map[string]string{
		"error": message,
	})
	return events.APIGatewayProxyResponse{
		StatusCode: status,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Body: string(body),
	}
}
