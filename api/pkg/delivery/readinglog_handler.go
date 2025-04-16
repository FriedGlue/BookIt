// pkg/delivery/readinglog_handler.go
package delivery

import (
	"encoding/json"
	"net/http"

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

func (h *ReadingLogHandler) UpdateReadingLogItem(w http.ResponseWriter, r *http.Request) {
	userID := extractUserID(r)
	var payload struct {
		EntryID   string `json:"readingLogItemId"`
		PagesRead int    `json:"pagesRead"`
		Notes     string `json:"notes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	if err := h.svc.Update(r.Context(), userID, payload.EntryID, payload.PagesRead, payload.Notes); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *ReadingLogHandler) DeleteReadingLogItem(w http.ResponseWriter, r *http.Request) {
	userID := extractUserID(r)
	entryID := r.URL.Query().Get("readingLogId")
	if err := h.svc.Delete(r.Context(), userID, entryID); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
