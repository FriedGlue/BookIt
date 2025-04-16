// pkg/usecase/readinglog.go
package usecase

import (
	"context"

	"github.com/FriedGlue/BookIt/api/pkg/models"
)

// ReadingLogService handles log entries.
type ReadingLogService interface {
	List(ctx context.Context, userID string) ([]models.ReadingLogItem, error)
	Create(ctx context.Context, userID string, entry models.ReadingLogItem) error
	Update(ctx context.Context, userID, entryID string, pagesRead int, notes string) error
	Delete(ctx context.Context, userID, entryID string) error
}
