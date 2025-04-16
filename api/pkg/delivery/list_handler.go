// pkg/delivery/list_handler.go
package delivery

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/FriedGlue/BookIt/api/pkg/models"
	"github.com/FriedGlue/BookIt/api/pkg/usecase"
)

// ReadingListHandler handles user bookshelf endpoints (toBeRead, read, custom).
type ReadingListHandler struct {
	svc usecase.ReadingListService
}

// NewReadingListHandler creates a ReadingListHandler.
func NewReadingListHandler(s usecase.ReadingListService) *ReadingListHandler {
	return &ReadingListHandler{svc: s}
}

// GetLists handles GET /list to retrieve all user's lists
func (h *ReadingListHandler) GetLists(w http.ResponseWriter, r *http.Request) {
	userID := extractUserID(r)
	if userID == "" {
		http.Error(w, "User ID not found in request", http.StatusUnauthorized)
		return
	}

	lists, err := h.svc.GetLists(r.Context(), userID)
	if err != nil {
		log.Printf("Error retrieving lists for user %s: %v", userID, err)
		http.Error(w, "Error retrieving lists: "+err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(lists)
}

// AddToList handles POST /list
func (h *ReadingListHandler) AddToList(w http.ResponseWriter, r *http.Request) {
	userID := extractUserID(r)
	if userID == "" {
		http.Error(w, "User ID not found in request", http.StatusUnauthorized)
		return
	}

	var payload struct {
		ListType  string `json:"listType"`
		BookID    string `json:"bookId"`
		Rating    int    `json:"rating,omitempty"`
		Review    string `json:"review,omitempty"`
		Thumbnail string `json:"thumbnail,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	book := models.Book{
		BookID:    payload.BookID,
		Thumbnail: payload.Thumbnail,
	}

	var err error
	switch payload.ListType {
	case "toBeRead":
		err = h.svc.AddToBeRead(r.Context(), userID, book)
	case "read":
		err = h.svc.AddToRead(r.Context(), userID, book, payload.Rating, payload.Review)
	default:
		err = h.svc.AddToCustomList(r.Context(), userID, payload.ListType, book)
	}

	if err != nil {
		log.Printf("Error adding book to list for user %s: %v", userID, err)
		http.Error(w, "Error adding to list: "+err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Book added to list successfully",
	})
}

// DeleteList handles DELETE /list?listName=
func (h *ReadingListHandler) DeleteList(w http.ResponseWriter, r *http.Request) {
	userID := extractUserID(r)
	if userID == "" {
		http.Error(w, "User ID not found in request", http.StatusUnauthorized)
		return
	}

	listName := r.URL.Query().Get("listName")
	if listName == "" {
		http.Error(w, "listName query parameter is required", http.StatusBadRequest)
		return
	}

	if err := h.svc.DeleteCustomList(r.Context(), userID, listName); err != nil {
		log.Printf("Error deleting list for user %s: %v", userID, err)
		http.Error(w, "Error deleting list: "+err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// DeleteListItem handles DELETE /list?listType=&bookId=
func (h *ReadingListHandler) DeleteListItem(w http.ResponseWriter, r *http.Request) {
	userID := extractUserID(r)
	if userID == "" {
		http.Error(w, "User ID not found in request", http.StatusUnauthorized)
		return
	}

	listType := r.URL.Query().Get("listType")
	bookID := r.URL.Query().Get("bookId")

	if listType == "" || bookID == "" {
		http.Error(w, "listType and bookId query parameters are required", http.StatusBadRequest)
		return
	}

	if err := h.svc.RemoveFromList(r.Context(), userID, listType, bookID); err != nil {
		log.Printf("Error removing book from list for user %s: %v", userID, err)
		http.Error(w, "Error removing from list: "+err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
