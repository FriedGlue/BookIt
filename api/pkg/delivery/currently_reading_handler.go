// pkg/delivery/currently_reading_handler.go
package delivery

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/FriedGlue/BookIt/api/pkg/models"
	"github.com/FriedGlue/BookIt/api/pkg/usecase"
)

// CurrentlyReadingHandler handles currently-reading endpoints.
type CurrentlyReadingHandler struct {
	svc usecase.CurrentlyReadingService
}

// NewCurrentlyReadingHandler creates a CurrentlyReadingHandler.
func NewCurrentlyReadingHandler(s usecase.CurrentlyReadingService) *CurrentlyReadingHandler {
	return &CurrentlyReadingHandler{svc: s}
}

func (h *CurrentlyReadingHandler) GetCurrentlyReading(w http.ResponseWriter, r *http.Request) {
	userID := extractUserID(r)
	if userID == "" {
		http.Error(w, "User ID not found in request", http.StatusUnauthorized)
		return
	}

	list, err := h.svc.GetCurrentlyReading(r.Context(), userID)
	if err != nil {
		log.Printf("Error retrieving currently reading for user %s: %v", userID, err)
		http.Error(w, "Error retrieving currently reading: "+err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(list)
}

func (h *CurrentlyReadingHandler) AddToCurrentlyReading(w http.ResponseWriter, r *http.Request) {
	userID := extractUserID(r)
	if userID == "" {
		http.Error(w, "User ID not found in request", http.StatusUnauthorized)
		return
	}

	var payload struct {
		BookID    string   `json:"bookId"`
		ISBN      string   `json:"isbn"`
		Title     string   `json:"title,omitempty"`
		Authors   []string `json:"authors,omitempty"`
		Thumbnail string   `json:"thumbnail,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	book := models.Book{
		BookID:    payload.BookID,
		ISBN:      payload.ISBN,
		Title:     payload.Title,
		Authors:   payload.Authors,
		Thumbnail: payload.Thumbnail,
	}

	if err := h.svc.AddToCurrentlyReading(r.Context(), userID, book); err != nil {
		log.Printf("Error adding book to currently reading for user %s: %v", userID, err)
		http.Error(w, "Error adding to currently reading: "+err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Book added to currently reading successfully",
	})
}

func (h *CurrentlyReadingHandler) UpdateCurrentlyReading(w http.ResponseWriter, r *http.Request) {
	userID := extractUserID(r)
	if userID == "" {
		http.Error(w, "User ID not found in request", http.StatusUnauthorized)
		return
	}

	var payload struct {
		BookID      string `json:"bookId"`
		CurrentPage int    `json:"currentPage"`
		Notes       string `json:"notes"`
		Date        string `json:"date"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	if payload.Date == "" {
		payload.Date = time.Now().Format(time.RFC3339)
	}

	if err := h.svc.UpdateProgress(r.Context(), userID, payload.BookID, payload.CurrentPage, payload.Notes, payload.Date); err != nil {
		log.Printf("Error updating reading progress for user %s, book %s: %v", userID, payload.BookID, err)
		http.Error(w, "Error updating progress: "+err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Reading progress updated successfully",
	})
}

func (h *CurrentlyReadingHandler) RemoveFromCurrentlyReading(w http.ResponseWriter, r *http.Request) {
	userID := extractUserID(r)
	if userID == "" {
		http.Error(w, "User ID not found in request", http.StatusUnauthorized)
		return
	}

	bookID := r.URL.Query().Get("bookId")
	if bookID == "" {
		http.Error(w, "bookId query parameter is required", http.StatusBadRequest)
		return
	}

	if err := h.svc.RemoveFromCurrentlyReading(r.Context(), userID, bookID); err != nil {
		log.Printf("Error removing book from currently reading for user %s, book %s: %v", userID, bookID, err)
		http.Error(w, "Error removing from currently reading: "+err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
