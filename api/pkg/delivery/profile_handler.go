// pkg/delivery/profile_handler.go
package delivery

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/FriedGlue/BookIt/api/pkg/usecase"
)

// ProfileHandler handles profile management endpoints.
type ProfileHandler struct {
	svc usecase.ProfileService
}

// NewProfileHandler creates a ProfileHandler.
func NewProfileHandler(s usecase.ProfileService) *ProfileHandler {
	return &ProfileHandler{svc: s}
}

// GetProfile handles GET /profile to retrieve user profile data
func (h *ProfileHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	userID := extractUserID(r)
	if userID == "" {
		http.Error(w, "User ID not found in request", http.StatusUnauthorized)
		return
	}

	// Direct access to profile through repository
	profile, err := h.svc.GetProfile(r.Context(), userID)
	if err != nil {
		log.Printf("Error retrieving profile for user %s: %v", userID, err)
		http.Error(w, "Error retrieving profile: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Return the full profile to the client
	json.NewEncoder(w).Encode(profile)
}

// UpdateProfile handles PUT /profile to update user profile information
func (h *ProfileHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	userID := extractUserID(r)
	if userID == "" {
		http.Error(w, "User ID not found in request", http.StatusUnauthorized)
		return
	}

	// Get the current profile first
	profile, err := h.svc.GetProfile(r.Context(), userID)
	if err != nil {
		log.Printf("Error retrieving profile for user %s: %v", userID, err)
		http.Error(w, "Error retrieving profile: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Parse update request
	var payload struct {
		Username string `json:"username"`
		Email    string `json:"email"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	// Update profile information
	if payload.Username != "" {
		profile.ProfileInformation.Username = payload.Username
	}
	if payload.Email != "" {
		profile.ProfileInformation.Email = payload.Email
	}

	// Save the updated profile
	if err := h.svc.SaveProfile(r.Context(), profile); err != nil {
		log.Printf("Error saving profile for user %s: %v", userID, err)
		http.Error(w, "Error saving profile: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Profile updated successfully",
	})
}
