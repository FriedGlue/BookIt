// pkg/delivery/auth.go
package delivery

import (
	"net/http"

	"github.com/awslabs/aws-lambda-go-api-proxy/core"
)

// extractUserID retrieves the Cognito 'sub' from API Gateway authorizer claims.
func extractUserID(r *http.Request) string {
	ctx := r.Context()
	apiGwCtx, ok := core.GetAPIGatewayContextFromContext(ctx)
	if !ok {
		return ""
	}
	claims, ok := apiGwCtx.Authorizer["claims"].(map[string]interface{})
	if !ok {
		return ""
	}
	sub, ok := claims["sub"].(string)
	if !ok || sub == "" {
		return ""
	}
	return sub
}
