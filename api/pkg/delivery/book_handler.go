// pkg/delivery/book_handler.go
package delivery

import (
	"encoding/json"
	"net/http"

	"github.com/FriedGlue/BookIt/api/pkg/usecase"
	"github.com/go-chi/chi/v5"
)

// BookHandler handles book-related endpoints.
type BookHandler struct {
	svc usecase.BookService
}

// NewBookHandler creates a BookHandler.
func NewBookHandler(s usecase.BookService) *BookHandler {
	return &BookHandler{svc: s}
}

func (h *BookHandler) GetBooks(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	bookId := chi.URLParam(r, "bookId")
	if bookId != "" {
		book, err := h.svc.Get(ctx, bookId)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		json.NewEncoder(w).Encode(book)
	} else {
		books, err := h.svc.List(ctx)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(books)
	}
}

func (h *BookHandler) CreateBook(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		ISBN string `json:"isbn"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	book, err := h.svc.CreateByISBN(r.Context(), payload.ISBN)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(book)
}

func (h *BookHandler) UpdateBook(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "bookId")
	var fields map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&fields); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	book, err := h.svc.Update(r.Context(), id, fields)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(book)
}

func (h *BookHandler) DeleteBook(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "bookId")
	if err := h.svc.Delete(r.Context(), id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
