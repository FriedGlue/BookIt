// pkg/usecase/profile.go
package usecase

import (
	"context"

	"github.com/FriedGlue/BookIt/api/pkg/models"
)

// ProfileService provides basic profile data operations
type ProfileService interface {
	// Core profile operations
	GetProfile(ctx context.Context, userID string) (models.Profile, error)
	SaveProfile(ctx context.Context, profile models.Profile) error
}
