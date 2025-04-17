// pkg/service/book_svc.go
package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/FriedGlue/BookIt/api/pkg/models"
	"github.com/FriedGlue/BookIt/api/pkg/usecase"
)

type bookSvc struct {
	repo usecase.BookRepo
}

// NewBookService constructs BookService.
func NewBookService(r usecase.BookRepo) usecase.BookService {
	return &bookSvc{repo: r}
}

func (s *bookSvc) Get(ctx context.Context, id string) (models.Book, error) {
	return s.repo.Load(ctx, id)
}

func (s *bookSvc) List(ctx context.Context) ([]models.Book, error) {
	return s.repo.QueryAll(ctx)
}

func (s *bookSvc) CreateByISBN(ctx context.Context, isbn string) (models.Book, error) {
	// try existing
	if books, _ := s.repo.SearchByISBN(ctx, isbn); len(books) > 0 {
		return books[0], nil
	}
	// fetch from OpenLibrary
	url := fmt.Sprintf("https://openlibrary.org/api/books?bibkeys=ISBN:%s&format=json&jscmd=data", isbn)
	resp, err := http.Get(url)
	if err != nil {
		return models.Book{}, err
	}
	defer resp.Body.Close()
	var data map[string]struct {
		Title         string                  `json:"title"`
		Authors       []struct{ Name string } `json:"authors"`
		NumberOfPages int                     `json:"number_of_pages"`
		Cover         struct{ Large string }  `json:"cover"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return models.Book{}, err
	}
	key := "ISBN:" + isbn
	ol := data[key]
	authors := []string{}
	for _, a := range ol.Authors {
		authors = append(authors, a.Name)
	}
	book := models.Book{
		BookID:     fmt.Sprintf("%d", time.Now().UnixNano()),
		ISBN:       isbn,
		Title:      ol.Title,
		Authors:    authors,
		TotalPages: ol.NumberOfPages,
		Thumbnail:  ol.Cover.Large,
	}
	if err := s.repo.Save(ctx, book); err != nil {
		return models.Book{}, err
	}
	return book, nil
}

func (s *bookSvc) Update(ctx context.Context, id string, fields map[string]interface{}) (models.Book, error) {
	b, err := s.repo.Load(ctx, id)
	if err != nil {
		return b, err
	}
	// merge fields (manual or via reflection)
	if title, ok := fields["title"].(string); ok {
		b.Title = title
	}
	if err := s.repo.Save(ctx, b); err != nil {
		return b, err
	}
	return b, nil
}

func (s *bookSvc) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

func (s *bookSvc) Search(ctx context.Context, params map[string]string) ([]models.Book, error) {
	log.Printf("Book service Search called with params: %v", params)

	if isbn := params["isbn"]; isbn != "" {
		log.Printf("Searching by ISBN: %s", isbn)
		books, err := s.repo.SearchByISBN(ctx, isbn)
		log.Printf("ISBN search returned %d books, error: %v", len(books), err)
		return books, err
	}
	if title := params["title"]; title != "" {
		log.Printf("Searching by title: %s", title)
		books, err := s.repo.SearchByTitle(ctx, title)
		log.Printf("Title search returned %d books, error: %v", len(books), err)
		return books, err
	}

	log.Printf("No search parameters provided")
	return nil, nil
}
