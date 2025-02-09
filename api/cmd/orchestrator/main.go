package main

import (
	"context"
	"log"
	"net/http"
	"strings"

	"github.com/FriedGlue/BookIt/api/pkg/handlers"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

// addCORSHeaders adds the appropriate headers to allow cross-origin requests.
func addCORSHeaders(response events.APIGatewayProxyResponse) events.APIGatewayProxyResponse {
	if response.Headers == nil {
		response.Headers = map[string]string{}
	}
	// response.Headers["Access-Control-Allow-Origin"] = "http://localhost:5173"
	response.Headers["Access-Control-Allow-Origin"] = "https://getbookit.org"
	response.Headers["Access-Control-Allow-Methods"] = "GET, POST, PUT, DELETE, OPTIONS"
	response.Headers["Access-Control-Allow-Headers"] = "Content-Type, Authorization, X-Amz-Date, X-Api-Key, X-Amz-Security-Token"
	response.Headers["Access-Control-Allow-Credentials"] = "TRUE"
	response.Headers["Content-Type"] = "application/json"

	return response
}

// handler is the main entry point for the Lambda.
func handler(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	var response events.APIGatewayProxyResponse

	path := request.Path
	method := request.HTTPMethod

	switch {
	case path == "/books/search" && method == "GET":
		response = handlers.SearchBooks(request)

	case strings.HasPrefix(path, "/books"):
		// Handle /books routes
		switch method {
		case "GET":
			response = handlers.GetBooks(request)
		case "POST":
			response = handlers.CreateBook(request)
		case "PUT":
			response = handlers.UpdateBook(request)
		case "DELETE":
			response = handlers.DeleteBook(request)
		default:
			response = events.APIGatewayProxyResponse{
				StatusCode: 405,
				Body:       "Method Not Allowed for /books",
			}
		}

	case strings.HasPrefix(path, "/currently-reading/start-reading"):
		response = handlers.StartReading(request)
	case strings.HasPrefix(path, "/currently-reading/finish-reading"):
		response = handlers.FinishReading(request)

	case strings.HasPrefix(path, "/currently-reading"):
		// Handle /currently-reading routes
		switch method {
		case "GET":
			response = handlers.GetCurrentlyReading(request)
		case "POST":
			response = handlers.AddToCurrentlyReading(request)
		case "PUT":
			response = handlers.UpdateCurrentlyReading(request)
		case "DELETE":
			response = handlers.RemoveFromCurrentlyReading(request)
		default:
			response = events.APIGatewayProxyResponse{
				StatusCode: 405,
				Body:       "Method Not Allowed for /currently-reading",
			}
		}

	case strings.HasPrefix(path, "/list"):
		// Handle /list routes
		switch method {
		case "GET":
			response = handlers.GetList(request)
		case "POST":
			if request.QueryStringParameters["listName"] != "" {
				response = handlers.CreateCustomBookshelf(request)
			} else {
				response = handlers.AddToList(request)
			}
		case "PUT":
			response = handlers.UpdateListItem(request)
		case "DELETE":
			if request.QueryStringParameters["listName"] != "" {
				response = handlers.DeleteCustomBookshelf(request)
			} else {
				response = handlers.DeleteListItem(request)
			}
		default:
			response = events.APIGatewayProxyResponse{
				StatusCode: 405,
				Body:       "Method Not Allowed for /list",
			}
		}

	case strings.HasPrefix(path, "/profile"):
		// Handle /profile routes
		switch method {
		case "GET":
			response = handlers.GetProfile(request)
		case "POST", "PUT":
			response = handlers.CreateOrUpdateProfile(request)
		case "DELETE":
			response = handlers.DeleteProfile(request)
		default:
			response = events.APIGatewayProxyResponse{
				StatusCode: 405,
				Body:       "Method Not Allowed for /profile",
			}
		}

	case strings.HasPrefix(path, "/reading-log"):
		// Handle /reading-log routes
		switch method {
		case "GET":
			response = handlers.GetReadingLog(request)
		case "DELETE":
			response = handlers.DeleteReadingLogItem(request)
		case "PUT":
			response = handlers.UpdateReadingLogItem(request)
		default:
			response = events.APIGatewayProxyResponse{
				StatusCode: 405,
				Body:       "Method Not Allowed for /reading-log",
			}
		}

	case strings.HasPrefix(path, "/challenges/"):
		// Extract the ID from the path
		pathParts := strings.Split(path, "/")
		if len(pathParts) == 3 && pathParts[2] != "" {
			// This is a request for a specific challenge
			switch method {
			case http.MethodPut:
				response = handlers.UpdateChallenge(request)
			case http.MethodDelete:
				response = handlers.DeleteChallenge(request)
			default:
				response = events.APIGatewayProxyResponse{
					StatusCode: 405,
					Body:       "Method Not Allowed for /challenges/{id}",
				}
			}
		} else {
			response = events.APIGatewayProxyResponse{
				StatusCode: 404,
				Body:       "Not Found",
			}
		}

	case path == "/challenges":
		switch method {
		case http.MethodPost:
			response = handlers.CreateChallenge(request)
		case http.MethodGet:
			response = handlers.GetChallenges(request)
		default:
			response = events.APIGatewayProxyResponse{
				StatusCode: 405,
				Body:       "Method Not Allowed for /challenges",
			}
		}

	default:
		response = events.APIGatewayProxyResponse{
			StatusCode: 404,
			Body:       "Not Found",
		}
	}

	response = addCORSHeaders(response)
	log.Println(response)
	return response, nil
}

func main() {
	lambda.Start(handler)
}
