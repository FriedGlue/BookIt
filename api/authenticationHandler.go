package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
)

// We'll read these from environment variables or
// pass them in at deploy time:
var (
	userPoolID       = os.Getenv("USER_POOL_ID")
	userPoolClientID = os.Getenv("USER_POOL_CLIENT_ID")
)

// newCognitoClient initializes the AWS Cognito client.
func newCognitoClient() *cognitoidentityprovider.CognitoIdentityProvider {
	sess := session.Must(session.NewSession())
	return cognitoidentityprovider.New(sess)
}

// ----------------- Sign Up Logic -----------------
//
// POST /auth/signup
// Request body (JSON): { "username":"...", "password":"...", "email":"..." }
//

func handleSignUp(request events.APIGatewayProxyRequest) events.APIGatewayProxyResponse {
	var payload struct {
		Username string `json:"username"`
		Password string `json:"password"`
		Email    string `json:"email"`
	}
	if err := json.Unmarshal([]byte(request.Body), &payload); err != nil {
		return errorResponse(400, "Invalid JSON: "+err.Error())
	}
	if payload.Username == "" || payload.Password == "" || payload.Email == "" {
		return errorResponse(400, "username, password, and email are required")
	}

	cip := newCognitoClient()
	input := &cognitoidentityprovider.SignUpInput{
		ClientId: aws.String(userPoolClientID),
		Username: aws.String(payload.Username),
		Password: aws.String(payload.Password),
		UserAttributes: []*cognitoidentityprovider.AttributeType{
			{
				Name:  aws.String("email"),
				Value: aws.String(payload.Email),
			},
		},
	}

	_, err := cip.SignUp(input)
	if err != nil {
		return errorResponse(500, fmt.Sprintf("SignUp error: %v", err))
	}

	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Body:       fmt.Sprintf("User '%s' sign-up initiated. Check your email for a confirmation code.", payload.Username),
	}
}

// ----------------- Confirm Sign Up Logic -----------------
//
// POST /auth/confirm
// Request body (JSON): { "username":"...", "code":"123456" }
//

func handleConfirmSignUp(request events.APIGatewayProxyRequest) events.APIGatewayProxyResponse {
	var payload struct {
		Username string `json:"username"`
		Code     string `json:"code"`
	}
	if err := json.Unmarshal([]byte(request.Body), &payload); err != nil {
		return errorResponse(400, "Invalid JSON: "+err.Error())
	}
	if payload.Username == "" || payload.Code == "" {
		return errorResponse(400, "username and code are required")
	}

	cip := newCognitoClient()
	input := &cognitoidentityprovider.ConfirmSignUpInput{
		ClientId:         aws.String(userPoolClientID),
		Username:         aws.String(payload.Username),
		ConfirmationCode: aws.String(payload.Code),
	}

	_, err := cip.ConfirmSignUp(input)
	if err != nil {
		return errorResponse(500, fmt.Sprintf("ConfirmSignUp error: %v", err))
	}

	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Body:       fmt.Sprintf("User '%s' confirmed successfully.", payload.Username),
	}
}

// ----------------- Sign In Logic -----------------
//
// POST /auth/signin
// Request body (JSON): { "username":"...", "password":"..." }
//

func handleSignIn(request events.APIGatewayProxyRequest) events.APIGatewayProxyResponse {
	var payload struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := json.Unmarshal([]byte(request.Body), &payload); err != nil {
		return errorResponse(400, "Invalid JSON: "+err.Error())
	}
	if payload.Username == "" || payload.Password == "" {
		return errorResponse(400, "username and password are required")
	}

	cip := newCognitoClient()
	input := &cognitoidentityprovider.InitiateAuthInput{
		AuthFlow: aws.String("USER_PASSWORD_AUTH"), // or "ADMIN_USER_PASSWORD_AUTH" if you use admin creds
		ClientId: aws.String(userPoolClientID),
		AuthParameters: map[string]*string{
			"USERNAME": aws.String(payload.Username),
			"PASSWORD": aws.String(payload.Password),
		},
	}

	output, err := cip.InitiateAuth(input)
	if err != nil {
		return errorResponse(401, fmt.Sprintf("SignIn error: %v", err))
	}

	if output.AuthenticationResult == nil {
		return errorResponse(401, "No authentication result returned.")
	}

	// Build a JSON response with the tokens
	tokens := map[string]string{
		"IdToken":      aws.StringValue(output.AuthenticationResult.IdToken),
		"AccessToken":  aws.StringValue(output.AuthenticationResult.AccessToken),
		"RefreshToken": aws.StringValue(output.AuthenticationResult.RefreshToken),
	}

	body, _ := json.Marshal(tokens)
	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Body:       string(body),
	}
}
