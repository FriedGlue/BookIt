// pkg/service/currently_reading_svc.go
package service

import (
	"context"
	"errors"
	"time"

	"github.com/FriedGlue/BookIt/api/pkg/models"
	"github.com/FriedGlue/BookIt/api/pkg/usecase"
)

type currentlyReadingSvc struct {
	repo       usecase.ProfileRepo
	logService usecase.ReadingLogService
	bookRepo   usecase.BookRepo
}

// NewCurrentlyReadingService creates a new currently reading service
func NewCurrentlyReadingService(r usecase.ProfileRepo, log usecase.ReadingLogService, book usecase.BookRepo) usecase.CurrentlyReadingService {
	return &currentlyReadingSvc{repo: r, logService: log, bookRepo: book}
}

func (s *currentlyReadingSvc) GetCurrentlyReading(ctx context.Context, userID string) ([]models.CurrentlyReadingItem, error) {
	p, err := s.repo.LoadProfile(ctx, userID)
	if err != nil {
		return nil, err
	}
	return p.CurrentlyReading, nil
}

func (s *currentlyReadingSvc) AddToCurrentlyReading(ctx context.Context, userID string, book models.Book) error {
	p, err := s.repo.LoadProfile(ctx, userID)
	if err != nil {
		return err
	}
	item := models.CurrentlyReadingItem{
		Book: models.Book{
			BookID:     book.BookID,
			Title:      book.Title,
			Authors:    book.Authors,
			Thumbnail:  book.Thumbnail,
			TotalPages: book.TotalPages,
		},
		StartedDate: time.Now().Format(time.RFC3339),
	}
	p.CurrentlyReading = append(p.CurrentlyReading, item)
	logEntry := models.ReadingLogItem{
		BookID:        book.BookID,
		Title:         book.Title,
		BookThumbnail: book.Thumbnail,
		PagesRead:     0,
		Notes:         "Book Started",
		Date:          time.Now().Format(time.RFC3339),
	}
	if err := s.logService.Create(ctx, userID, logEntry); err != nil {
		return err
	}
	return s.repo.SaveProfile(ctx, p)
}

func (s *currentlyReadingSvc) UpdateProgress(ctx context.Context, userID, bookID string, currentPage int, notes string, date string) error {
	p, err := s.repo.LoadProfile(ctx, userID)
	if err != nil {
		return err
	}
	var lastRead int
	found := false
	for i := range p.CurrentlyReading {
		if p.CurrentlyReading[i].Book.BookID == bookID {
			lastRead = p.CurrentlyReading[i].Book.Progress.LastPageRead
			p.CurrentlyReading[i].Book.Progress.LastPageRead = currentPage
			p.CurrentlyReading[i].Book.Progress.Percentage = float64(currentPage) / float64(p.CurrentlyReading[i].Book.TotalPages) * 100
			p.CurrentlyReading[i].Book.Progress.LastUpdated = time.Now().Format(time.RFC3339)
			found = true
			break
		}
	}
	if !found {
		return errors.New("book not found in currently reading list")
	}

	logEntry := models.ReadingLogItem{
		BookID:    bookID,
		PagesRead: currentPage - lastRead,
		Notes:     notes,
		Date:      date,
	}
	if err := s.logService.Create(ctx, userID, logEntry); err != nil {
		return err
	}
	return s.repo.SaveProfile(ctx, p)
}

func (s *currentlyReadingSvc) RemoveFromCurrentlyReading(ctx context.Context, userID, bookID string) error {
	p, err := s.repo.LoadProfile(ctx, userID)
	if err != nil {
		return err
	}
	for i := range p.CurrentlyReading {
		if p.CurrentlyReading[i].Book.BookID == bookID {
			p.CurrentlyReading = append(p.CurrentlyReading[:i], p.CurrentlyReading[i+1:]...)
			break
		}
	}
	return s.repo.SaveProfile(ctx, p)
}

func (s *currentlyReadingSvc) StartReading(ctx context.Context, userID, bookID, fromList string) error {
	// Implementation details
	return nil
}

func (s *currentlyReadingSvc) FinishReading(ctx context.Context, userID, bookID string) error {
	// Implementation details
	return nil
}
