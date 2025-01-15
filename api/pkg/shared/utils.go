package shared

import (
	"fmt"
	"log"

	"github.com/aws/aws-lambda-go/events"
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
