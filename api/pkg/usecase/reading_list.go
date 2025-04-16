// pkg/usecase/reading_list.go
package usecase

import (
	"context"

	"github.com/FriedGlue/BookIt/api/pkg/models"
)

// ReadingListService manages all types of reading lists (standard and custom)
type ReadingListService interface {
	// Standard lists
	GetLists(ctx context.Context, userID string) (map[string][]models.ReadItem, error)
	GetToBeRead(ctx context.Context, userID string) ([]models.ReadItem, error)
	GetRead(ctx context.Context, userID string) ([]models.ReadItem, error)
	AddToBeRead(ctx context.Context, userID string, book models.Book) error
	AddToRead(ctx context.Context, userID string, book models.Book, rating int, review string) error

	// Custom lists
	GetCustomLists(ctx context.Context, userID string) (map[string][]models.ReadItem, error)
	CreateCustomList(ctx context.Context, userID, listName string) error
	AddToCustomList(ctx context.Context, userID, listName string, book models.Book) error
	DeleteCustomList(ctx context.Context, userID, listName string) error

	// Common operations
	RemoveFromList(ctx context.Context, userID, listName, bookID string) error
}
