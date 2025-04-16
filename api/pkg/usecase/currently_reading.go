// pkg/usecase/currently_reading.go
package usecase

import (
	"context"

	"github.com/FriedGlue/BookIt/api/pkg/models"
)

// CurrentlyReadingService manages currently reading functionality
type CurrentlyReadingService interface {
	GetCurrentlyReading(ctx context.Context, userID string) ([]models.CurrentlyReadingItem, error)
	AddToCurrentlyReading(ctx context.Context, userID string, book models.Book) error
	UpdateProgress(ctx context.Context, userID, bookID string, currentPage int, notes string) error
	RemoveFromCurrentlyReading(ctx context.Context, userID, bookID string) error
	StartReading(ctx context.Context, userID, bookID, fromList string) error
	FinishReading(ctx context.Context, userID, bookID string) error
}
