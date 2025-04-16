// pkg/delivery/challenge_handler.go
package delivery

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/FriedGlue/BookIt/api/pkg/models"
	"github.com/FriedGlue/BookIt/api/pkg/usecase"
	"github.com/go-chi/chi/v5"
)

// ChallengeHandler handles reading challenge endpoints.
type ChallengeHandler struct {
	svc usecase.ReadingChallengeService
}

// NewChallengeHandler creates a ChallengeHandler.
func NewChallengeHandler(s usecase.ReadingChallengeService) *ChallengeHandler {
	return &ChallengeHandler{svc: s}
}

// GetChallenges handles GET /challenges
func (h *ChallengeHandler) GetChallenges(w http.ResponseWriter, r *http.Request) {
	userID := extractUserID(r)
	if userID == "" {
		http.Error(w, "User ID not found in request", http.StatusUnauthorized)
		return
	}

	challenges, err := h.svc.GetChallenges(r.Context(), userID)
	if err != nil {
		log.Printf("Error retrieving challenges for user %s: %v", userID, err)
		http.Error(w, "Error retrieving challenges: "+err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(challenges)
}

// CreateChallenge handles POST /challenges
func (h *ChallengeHandler) CreateChallenge(w http.ResponseWriter, r *http.Request) {
	userID := extractUserID(r)
	if userID == "" {
		http.Error(w, "User ID not found in request", http.StatusUnauthorized)
		return
	}

	var challenge models.ReadingChallenge
	if err := json.NewDecoder(r.Body).Decode(&challenge); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	if err := h.svc.CreateChallenge(r.Context(), userID, challenge); err != nil {
		log.Printf("Error creating challenge for user %s: %v", userID, err)
		http.Error(w, "Error creating challenge: "+err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Challenge created successfully",
	})
}

// UpdateChallenge handles PUT /challenges/{id}
func (h *ChallengeHandler) UpdateChallenge(w http.ResponseWriter, r *http.Request) {
	userID := extractUserID(r)
	if userID == "" {
		http.Error(w, "User ID not found in request", http.StatusUnauthorized)
		return
	}

	challengeID := getPathParam(r, "id")
	if challengeID == "" {
		http.Error(w, "Challenge ID is required", http.StatusBadRequest)
		return
	}

	var updates map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	if err := h.svc.UpdateChallenge(r.Context(), userID, challengeID, updates); err != nil {
		log.Printf("Error updating challenge for user %s: %v", userID, err)
		http.Error(w, "Error updating challenge: "+err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Challenge updated successfully",
	})
}

// DeleteChallenge handles DELETE /challenges/{id}
func (h *ChallengeHandler) DeleteChallenge(w http.ResponseWriter, r *http.Request) {
	userID := extractUserID(r)
	if userID == "" {
		http.Error(w, "User ID not found in request", http.StatusUnauthorized)
		return
	}

	challengeID := getPathParam(r, "id")
	if challengeID == "" {
		http.Error(w, "Challenge ID is required", http.StatusBadRequest)
		return
	}

	if err := h.svc.DeleteChallenge(r.Context(), userID, challengeID); err != nil {
		log.Printf("Error deleting challenge for user %s: %v", userID, err)
		http.Error(w, "Error deleting challenge: "+err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Helper function to extract path parameters
func getPathParam(r *http.Request, param string) string {
	if param == "" {
		return ""
	}
	ctx := chi.RouteContext(r.Context())
	if ctx == nil {
		return ""
	}
	return ctx.URLParam(param)
}
