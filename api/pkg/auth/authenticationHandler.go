package auth

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
	"github.com/aws/aws-sdk-go/service/sns"
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

func newSNSClient() *sns.SNS {
	sess := session.Must(session.NewSession())
	return sns.New(sess)
}

// ----------------- Sign Up Logic -----------------
//
// POST /auth/signup
// Request body (JSON): { "username":"...", "password":"...", "email":"..." }
//

// SignUpPayload represents the expected JSON structure for sign-up requests
type SignUpPayload struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Email    string `json:"email"`
}

func HandleSignUp(request events.APIGatewayProxyRequest) events.APIGatewayProxyResponse {
	var payload SignUpPayload
	if err := json.Unmarshal([]byte(request.Body), &payload); err != nil {
		return ErrorResponse(400, "Invalid JSON: "+err.Error())
	}
	if payload.Username == "" || payload.Password == "" || payload.Email == "" {
		return ErrorResponse(400, "username, password, and email are required")
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
		return ErrorResponse(500, fmt.Sprintf("SignUp error: %v", err))
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

func HandleConfirmSignUp(request events.APIGatewayProxyRequest) events.APIGatewayProxyResponse {
	var payload struct {
		Username string `json:"username"`
		Code     string `json:"code"`
	}
	if err := json.Unmarshal([]byte(request.Body), &payload); err != nil {
		return ErrorResponse(400, "Invalid JSON: "+err.Error())
	}
	if payload.Username == "" || payload.Code == "" {
		return ErrorResponse(400, "username and code are required")
	}

	cip := newCognitoClient()
	input := &cognitoidentityprovider.ConfirmSignUpInput{
		ClientId:         aws.String(userPoolClientID),
		Username:         aws.String(payload.Username),
		ConfirmationCode: aws.String(payload.Code),
	}

	_, err := cip.ConfirmSignUp(input)
	if err != nil {
		return ErrorResponse(500, fmt.Sprintf("ConfirmSignUp error: %v", err))
	}

	// Get the user's sub from Cognito
	userInfo, err := cip.AdminGetUser(&cognitoidentityprovider.AdminGetUserInput{
		UserPoolId: aws.String(userPoolID),
		Username:   aws.String(payload.Username),
	})
	if err != nil {
		log.Printf("Error getting user info: %v", err)
		return ErrorResponse(500, fmt.Sprintf("Error getting user info: %v", err))
	}

	var sub string
	for _, attr := range userInfo.UserAttributes {
		if *attr.Name == "sub" {
			sub = *attr.Value
			break
		}
	}

	if sub == "" {
		log.Printf("User sub not found")
		return ErrorResponse(500, "User sub not found")
	}

	// Send a message to the service bus to create a new user profile
	svc := newSNSClient()
	message := map[string]string{
		"action":   "CREATE_USER_PROFILE",
		"username": payload.Username,
		"sub":      sub,
	}

	messageJSON, err := json.Marshal(message)
	if err != nil {
		log.Printf("Error marshalling SNS message: %v", err)
		// Don't return error to client since confirmation was successful
	} else {
		_, err = svc.Publish(&sns.PublishInput{
			TopicArn: aws.String(os.Getenv("USER_EVENTS_TOPIC_ARN")),
			Message:  aws.String(string(messageJSON)),
		})
		if err != nil {
			log.Printf("Error publishing to SNS: %v", err)
			// Don't return error to client since confirmation was successful
		}
	}

	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Body:       fmt.Sprintf("User '%s' confirmed successfully.", payload.Username),
	}
}

// ----------------- Resend Confirmation Code Logic -----------------
//
// POST /auth/resend-confirmation-code
// Request body (JSON): { "username":"..." }
//

func HandleResendConfirmationCode(request events.APIGatewayProxyRequest) events.APIGatewayProxyResponse {
	var payload struct {
		Username string `json:"username"`
	}
	if err := json.Unmarshal([]byte(request.Body), &payload); err != nil {
		return ErrorResponse(400, "Invalid JSON: "+err.Error())
	}
	if payload.Username == "" {
		return ErrorResponse(400, "username is required")
	}

	cip := newCognitoClient()
	input := &cognitoidentityprovider.ResendConfirmationCodeInput{
		ClientId: aws.String(userPoolClientID),
		Username: aws.String(payload.Username),
	}

	_, err := cip.ResendConfirmationCode(input)
	if err != nil {
		return ErrorResponse(500, fmt.Sprintf("ResendConfirmationCode error: %v", err))
	}

	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Body:       fmt.Sprintf("Confirmation code resent to '%s'.", payload.Username),
	}
}

// ----------------- Sign In Logic -----------------
//
// POST /auth/signin
// Request body (JSON): { "username":"...", "password":"..." }
//

func HandleSignIn(request events.APIGatewayProxyRequest) events.APIGatewayProxyResponse {
	var payload struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := json.Unmarshal([]byte(request.Body), &payload); err != nil {
		return ErrorResponse(400, "Invalid JSON: "+err.Error())
	}
	if payload.Username == "" || payload.Password == "" {
		return ErrorResponse(400, "username and password are required")
	}

	cip := newCognitoClient()
	input := &cognitoidentityprovider.InitiateAuthInput{
		AuthFlow: aws.String("USER_PASSWORD_AUTH"),
		ClientId: aws.String(userPoolClientID),
		AuthParameters: map[string]*string{
			"USERNAME": aws.String(payload.Username),
			"PASSWORD": aws.String(payload.Password),
		},
	}

	output, err := cip.InitiateAuth(input)
	if err != nil {
		return ErrorResponse(401, fmt.Sprintf("SignIn error: %v", err))
	}

	if output.AuthenticationResult == nil {
		return ErrorResponse(401, "No authentication result returned.")
	}

	// Build a JSON response with the tokens
	tokens := map[string]string{
		"IdToken":      aws.StringValue(output.AuthenticationResult.IdToken),
		"AccessToken":  aws.StringValue(output.AuthenticationResult.AccessToken),
		"RefreshToken": aws.StringValue(output.AuthenticationResult.RefreshToken),
	}

	body, _ := json.Marshal(tokens)

	// Create response with refresh token cookie
	response := events.APIGatewayProxyResponse{
		StatusCode: 200,
		Body:       string(body),
		Headers: map[string]string{
			"Set-Cookie": fmt.Sprintf("refreshToken=%s; Path=/; Max-Age=2592000; HttpOnly; Secure; SameSite=none",
				aws.StringValue(output.AuthenticationResult.RefreshToken)),
		},
	}

	return response
}

// ----------------- Sign Out Logic -----------------
//
// POST /auth/signout
//

func HandleSignOut(request events.APIGatewayProxyRequest) events.APIGatewayProxyResponse {
	// Return response that clears the refresh token cookie
	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Headers: map[string]string{
			"Set-Cookie": "refreshToken=; Path=/; Expires=Thu, 01 Jan 1970 00:00:01 GMT; HttpOnly; Secure; SameSite=none",
		},
		Body: "Signed out successfully",
	}
}

// ----------------- Refresh Token Logic -----------------
//
// POST /auth/refresh
// Headers: { "Authorization": "Bearer <refresh_token>" }
//

func HandleRefresh(request events.APIGatewayProxyRequest) events.APIGatewayProxyResponse {
	// Get refresh token from Authorization header
	authHeader := request.Headers["Authorization"]
	if authHeader == "" {
		return ErrorResponse(401, "No Authorization header provided")
	}

	// Extract token from "Bearer <token>"
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return ErrorResponse(401, "Invalid Authorization header format")
	}
	refreshToken := parts[1]

	cip := newCognitoClient()
	input := &cognitoidentityprovider.InitiateAuthInput{
		AuthFlow: aws.String("REFRESH_TOKEN_AUTH"),
		ClientId: aws.String(userPoolClientID),
		AuthParameters: map[string]*string{
			"REFRESH_TOKEN": aws.String(refreshToken),
		},
	}

	output, err := cip.InitiateAuth(input)
	if err != nil {
		return ErrorResponse(401, fmt.Sprintf("Refresh token error: %v", err))
	}

	if output.AuthenticationResult == nil {
		return ErrorResponse(401, "No authentication result returned")
	}

	// Build a JSON response with the new tokens
	tokens := map[string]string{
		"IdToken":     aws.StringValue(output.AuthenticationResult.IdToken),
		"AccessToken": aws.StringValue(output.AuthenticationResult.AccessToken),
	}

	body, _ := json.Marshal(tokens)
	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Body:       string(body),
	}
}

// ErrorResponse is a helper to generate an APIGatewayProxyResponse with a given status and message.
func ErrorResponse(status int, message string) events.APIGatewayProxyResponse {
	body, _ := json.Marshal(map[string]string{
		"error": message,
	})
	return events.APIGatewayProxyResponse{
		StatusCode: status,
		Headers: map[string]string{
			"Content-Type":                 "application/json",
			"Access-Control-Allow-Origin":  "*",
			"Access-Control-Allow-Methods": "OPTIONS,POST,GET,PUT,DELETE",
			"Access-Control-Allow-Headers": "Content-Type,Authorization",
		},
		Body: string(body),
	}
}

// SuccessResponse is a helper to generate an APIGatewayProxyResponse with a given status and body.
func SuccessResponse(status int, body interface{}) events.APIGatewayProxyResponse {
	jsonBody, _ := json.Marshal(body)
	return events.APIGatewayProxyResponse{
		StatusCode: status,
		Headers: map[string]string{
			"Content-Type":                 "application/json",
			"Access-Control-Allow-Origin":  "*",
			"Access-Control-Allow-Methods": "OPTIONS,POST,GET,PUT,DELETE",
			"Access-Control-Allow-Headers": "Content-Type,Authorization",
		},
		Body: string(jsonBody),
	}
}
