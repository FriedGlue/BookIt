package main

import (
	"context"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

// addCORSHeaders adds the appropriate headers to allow cross-origin requests.
func addCORSHeaders(response events.APIGatewayProxyResponse) events.APIGatewayProxyResponse {
	if response.Headers == nil {
		response.Headers = map[string]string{}
	}
	response.Headers["Access-Control-Allow-Origin"] = "*"
	response.Headers["Access-Control-Allow-Methods"] = "GET, POST, PUT, DELETE, OPTIONS"
	response.Headers["Access-Control-Allow-Headers"] = "Content-Type"
	return response
}

// handler is the main entry point for the Lambda.
func handler(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	var response events.APIGatewayProxyResponse

	path := request.Path
	method := request.HTTPMethod

	switch {
	case strings.HasPrefix(path, "/books"):
		// Handle /books routes
		switch method {
		case "GET":
			response = GetBooks(request)
		case "POST":
			response = CreateBook(request)
		case "PUT":
			response = UpdateBook(request)
		case "DELETE":
			response = DeleteBook(request)
		default:
			response = events.APIGatewayProxyResponse{
				StatusCode: 405,
				Body:       "Method Not Allowed for /books",
			}
		}

	case strings.HasPrefix(path, "/lists"):
		// Handle /lists routes
		switch method {
		case "GET":
			response = GetUserLists(request)
		case "POST":
			response = CreateUserList(request)
		// Add others like PUT, DELETE if you need them for /lists
		default:
			response = events.APIGatewayProxyResponse{
				StatusCode: 405,
				Body:       "Method Not Allowed for /lists",
			}
		}

	case strings.HasPrefix(path, "/auth"):
		// Handle /auth routes
		// We'll match specific endpoints: /auth/signup, /auth/confirm, /auth/signin
		switch {
		case path == "/auth/signup" && method == "POST":
			response = handleSignUp(request)
		case path == "/auth/confirm" && method == "POST":
			response = handleConfirmSignUp(request)
		case path == "/auth/signin" && method == "POST":
			response = handleSignIn(request)
		default:
			response = events.APIGatewayProxyResponse{
				StatusCode: 405,
				Body:       "Method Not Allowed for /auth",
			}
		}

	default:
		response = events.APIGatewayProxyResponse{
			StatusCode: 404,
			Body:       "Not Found",
		}
	}

	return addCORSHeaders(response), nil
}

func main() {
	lambda.Start(handler)
}
