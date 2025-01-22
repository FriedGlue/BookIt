package main

import (
	"context"
	"log"
	"strings"

	"github.com/FriedGlue/BookIt/api/pkg/auth"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

// addCORSHeaders adds the appropriate headers to allow cross-origin requests.
func addCORSHeaders(response events.APIGatewayProxyResponse) events.APIGatewayProxyResponse {
	if response.Headers == nil {
		response.Headers = map[string]string{}
	}
	response.Headers["Access-Control-Allow-Origin"] = "http://localhost:5173"
	response.Headers["Access-Control-Allow-Methods"] = "GET, POST, PUT, DELETE, OPTIONS"
	response.Headers["Access-Control-Allow-Headers"] = "Content-Type, Authorization, X-Amz-Date, X-Api-Key, X-Amz-Security-Token"
	response.Headers["Access-Control-Allow-Credentials"] = "TRUE"
	response.Headers["Content-Type"] = "application/json"
	return response
}

// handler is the main entry point for the Lambda.
func handler(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// Log the incoming request details
	log.Printf("Received request: Method=%s, Path=%s\n", request.HTTPMethod, request.Path)

	var response events.APIGatewayProxyResponse

	path := request.Path
	method := request.HTTPMethod

	switch {
	case strings.HasPrefix(path, "/auth"):
		log.Printf("Handling /auth route: Method=%s, Path=%s\n", method, path)
		// Handle /auth routes
		// We'll match specific endpoints: /auth/signup, /auth/confirm, /auth/signin
		switch {
		case path == "/auth/signup" && method == "POST":
			log.Println("Invoking HandleSignUp")
			response = auth.HandleSignUp(request)

		case path == "/auth/confirm" && method == "POST":
			log.Println("Invoking HandleConfirmSignUp")
			response = auth.HandleConfirmSignUp(request)

		case path == "/auth/signin" && method == "POST":
			log.Println("Invoking HandleSignIn")
			response = auth.HandleSignIn(request)

		case path == "/auth/refresh" && method == "POST":
			log.Println("Invoking HandleRefresh")
			response = auth.HandleRefresh(request)

		case path == "/auth/signout" && method == "POST":
			log.Println("Invoking HandleSignOut")
			response = auth.HandleSignOut(request)

		case method == "OPTIONS":
			log.Println("Handling OPTIONS (preflight) for path:", path)
			// Return a 200 with no body, but addCORSHeaders will add the necessary CORS headers
			response = events.APIGatewayProxyResponse{
				StatusCode: 200,
				Body:       "",
			}

		default:
			log.Printf("Method not allowed on /auth: Method=%s, Path=%s\n", method, path)
			response = events.APIGatewayProxyResponse{
				StatusCode: 405,
				Body:       "Method Not Allowed for /auth",
			}
		}

	default:
		log.Printf("No route found: Method=%s, Path=%s\n", method, path)
		response = events.APIGatewayProxyResponse{
			StatusCode: 404,
			Body:       "Not Found",
		}
	}

	response = addCORSHeaders(response)
	log.Printf("Response: StatusCode=%d, Body=%s\n", response.StatusCode, response.Body)
	return response, nil
}

func main() {
	log.Println("Starting Lambda function")
	lambda.Start(handler)
}
