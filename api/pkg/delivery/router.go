// pkg/delivery/router.go
package delivery

import (
	"context"
	"log"
	"net/http"

	"github.com/FriedGlue/BookIt/api/pkg/usecase"
	"github.com/aws/aws-lambda-go/events"
	awschi "github.com/awslabs/aws-lambda-go-api-proxy/chi"
	"github.com/go-chi/chi/v5"
)

// Router wires HTTP endpoints to service methods.
type Router struct {
	BookSvc             usecase.BookService
	ProfileSvc          usecase.ProfileService
	LogSvc              usecase.ReadingLogService
	CurrentlyReadingSvc usecase.CurrentlyReadingService
	ReadingListSvc      usecase.ReadingListService
	ChallengeSvc        usecase.ReadingChallengeService
	adapter             *awschi.ChiLambda
}

// NewRouter constructs a Router with all routes registered.
func NewRouter(
	b usecase.BookService,
	p usecase.ProfileService,
	l usecase.ReadingLogService,
	cr usecase.CurrentlyReadingService,
	rl usecase.ReadingListService,
	ch usecase.ReadingChallengeService,
) *Router {
	r := chi.NewRouter()

	// Profile routes
	profileHandler := NewProfileHandler(p)
	r.Get("/profile", profileHandler.GetProfile)
	r.Put("/profile", profileHandler.UpdateProfile)

	// Books CRUD
	bookHandler := NewBookHandler(b)
	r.Post("/books", bookHandler.CreateBook)
	r.Get("/books", bookHandler.GetBooks)
	r.Get("/books/{bookId}", bookHandler.GetBooks)
	r.Put("/books/{bookId}", bookHandler.UpdateBook)
	r.Delete("/books/{bookId}", bookHandler.DeleteBook)

	// Book Search
	r.Get("/books/search", bookHandler.SearchBooks)

	// Currently Reading
	crHandler := NewCurrentlyReadingHandler(cr)
	r.Post("/currently-reading", crHandler.AddToCurrentlyReading)
	r.Get("/currently-reading", crHandler.GetCurrentlyReading)
	r.Put("/currently-reading", crHandler.UpdateCurrentlyReading)
	r.Delete("/currently-reading", crHandler.RemoveFromCurrentlyReading)

	// Reading Lists
	listHandler := NewReadingListHandler(rl)
	r.Get("/list", listHandler.GetLists)
	r.Post("/list", listHandler.AddToList)
	r.With(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Query().Get("listName") != "" {
				listHandler.DeleteList(w, r)
				return
			}
			listHandler.DeleteListItem(w, r)
		})
	}).Delete("/list", nil)

	// Reading Challenges
	challengeHandler := NewChallengeHandler(ch)
	r.Post("/challenges", challengeHandler.CreateChallenge)
	r.Get("/challenges", challengeHandler.GetChallenges)
	r.Put("/challenges/{id}", challengeHandler.UpdateChallenge)
	r.Delete("/challenges/{id}", challengeHandler.DeleteChallenge)

	// Reading Log
	logHandler := NewReadingLogHandler(l)
	r.Post("/reading-log", logHandler.CreateReadingLogItem)
	r.Get("/reading-log", logHandler.ListReadingLog)
	r.Put("/reading-log", logHandler.UpdateReadingLogItem)
	r.Delete("/reading-log", logHandler.DeleteReadingLogItem)

	return &Router{
		BookSvc:             b,
		ProfileSvc:          p,
		LogSvc:              l,
		CurrentlyReadingSvc: cr,
		ReadingListSvc:      rl,
		ChallengeSvc:        ch,
		adapter:             awschi.New(r),
	}
}

// Handler is the entrypoint for AWS Lambda.
func (rt *Router) Handler(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	log.Printf("Received request: Method=%s, Path=%s, Resource=%s", req.HTTPMethod, req.Path, req.Resource)

	// Log headers to check if Authorization is present
	for k, v := range req.Headers {
		if k == "Authorization" || k == "authorization" {
			log.Printf("Found %s header: %s...", k, v[:10]) // Log just the beginning to avoid exposing full token
		}
	}

	// Log query parameters
	if len(req.QueryStringParameters) > 0 {
		log.Printf("Query parameters: %v", req.QueryStringParameters)
	}

	// Call the adapter and log the response
	resp, err := rt.adapter.ProxyWithContext(ctx, req)
	if err != nil {
		log.Printf("Error processing request: %v", err)
	} else {
		log.Printf("Response: StatusCode=%d, Body length=%d", resp.StatusCode, len(resp.Body))
	}

	return resp, err
}
