// pkg/usecase/book.go
package usecase

import (
	"context"

	"github.com/FriedGlue/BookIt/api/pkg/models"
)

// BookService defines CRUD for books.
type BookService interface {
	Get(ctx context.Context, id string) (models.Book, error)
	List(ctx context.Context) ([]models.Book, error)
	CreateByISBN(ctx context.Context, isbn string) (models.Book, error)
	Update(ctx context.Context, id string, fields map[string]interface{}) (models.Book, error)
	Delete(ctx context.Context, id string) error
	Search(ctx context.Context, params map[string]string) ([]models.Book, error)
}
