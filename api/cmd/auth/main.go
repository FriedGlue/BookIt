package main

import (
	"context"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/friedglue/BookIt/api/pkg/auth"
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
	case strings.HasPrefix(path, "/auth"):
		// Handle /auth routes
		// We'll match specific endpoints: /auth/signup, /auth/confirm, /auth/signin
		switch {
		case path == "/auth/signup" && method == "POST":
			response = auth.HandleSignUp(request)
		case path == "/auth/confirm" && method == "POST":
			response = auth.HandleConfirmSignUp(request)
		case path == "/auth/signin" && method == "POST":
			response = auth.HandleSignIn(request)
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
