// pkg/delivery/readinglog_handler.go
package delivery

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/FriedGlue/BookIt/api/pkg/models"
	"github.com/FriedGlue/BookIt/api/pkg/usecase"
)

// ReadingLogHandler handles reading-log endpoints.
type ReadingLogHandler struct {
	svc usecase.ReadingLogService
}

// NewReadingLogHandler creates a ReadingLogHandler.
func NewReadingLogHandler(s usecase.ReadingLogService) *ReadingLogHandler {
	return &ReadingLogHandler{svc: s}
}

func (h *ReadingLogHandler) ListReadingLog(w http.ResponseWriter, r *http.Request) {
	userID := extractUserID(r)
	logEntries, err := h.svc.List(r.Context(), userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(logEntries)
}

func (h *ReadingLogHandler) CreateReadingLogItem(w http.ResponseWriter, r *http.Request) {
	userID := extractUserID(r)
	var payload struct {
		BookID    string `json:"bookId"`
		PagesRead int    `json:"pagesRead"`
		Notes     string `json:"notes"`
		Date      string `json:"date"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	if payload.Date == "" {
		payload.Date = time.Now().Format(time.RFC3339)
	}
	entry := models.ReadingLogItem{
		BookID:    payload.BookID,
		PagesRead: payload.PagesRead,
		Notes:     payload.Notes,
		Date:      payload.Date,
	}
	if err := h.svc.Create(r.Context(), userID, entry); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Reading log item created successfully",
	})
}

func (h *ReadingLogHandler) UpdateReadingLogItem(w http.ResponseWriter, r *http.Request) {
	userID := extractUserID(r)
	if userID == "" {
		http.Error(w, "User ID not found in request", http.StatusUnauthorized)
		return
	}

	var payload struct {
		EntryID   string `json:"readingLogItemId"`
		PagesRead int    `json:"pagesRead"`
		Notes     string `json:"notes"`
		Date      string `json:"date"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	if payload.Date == "" {
		payload.Date = time.Now().Format(time.RFC3339)
	}
	if err := h.svc.Update(r.Context(), userID, payload.EntryID, payload.PagesRead, payload.Notes, payload.Date); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Reading log item updated successfully",
	})
}

func (h *ReadingLogHandler) DeleteReadingLogItem(w http.ResponseWriter, r *http.Request) {
	userID := extractUserID(r)
	if userID == "" {
		http.Error(w, "User ID not found in request", http.StatusUnauthorized)
		return
	}
	entryID := r.URL.Query().Get("readingLogId")
	if entryID == "" {
		http.Error(w, "readingLogId query parameter is required", http.StatusBadRequest)
		return
	}
	if err := h.svc.Delete(r.Context(), userID, entryID); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Reading log item deleted successfully",
	})
}
