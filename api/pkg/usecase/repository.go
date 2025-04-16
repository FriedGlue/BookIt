// pkg/usecase/repository.go
package usecase

import (
	"context"

	"github.com/FriedGlue/BookIt/api/pkg/models"
)

// BookRepo abstracts book persistence.
type BookRepo interface {
	Load(ctx context.Context, id string) (models.Book, error)
	QueryAll(ctx context.Context) ([]models.Book, error)
	Save(ctx context.Context, b models.Book) error
	Delete(ctx context.Context, id string) error
	SearchByISBN(ctx context.Context, isbn string) ([]models.Book, error)
	SearchByTitle(ctx context.Context, q string) ([]models.Book, error)
}

// ProfileRepo abstracts profile persistence.
type ProfileRepo interface {
	LoadProfile(ctx context.Context, userID string) (models.Profile, error)
	SaveProfile(ctx context.Context, p models.Profile) error
}
