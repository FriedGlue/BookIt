// pkg/usecase/profile.go
package usecase

import (
	"context"

	"github.com/FriedGlue/BookIt/api/pkg/models"
)

// ProfileService provides direct access to profile data
type ProfileService interface {
	// Direct profile access
	GetProfile(ctx context.Context, userID string) (models.Profile, error)
	SaveProfile(ctx context.Context, profile models.Profile) error
}

// ProfileServiceImpl implements ProfileService using a repository
type ProfileServiceImpl struct {
	repo ProfileRepo
}

// NewProfileService creates a new ProfileService
func NewProfileService(repo ProfileRepo) ProfileService {
	return &ProfileServiceImpl{
		repo: repo,
	}
}

// GetProfile directly loads profile data from repository
func (s *ProfileServiceImpl) GetProfile(ctx context.Context, userID string) (models.Profile, error) {
	return s.repo.LoadProfile(ctx, userID)
}

// SaveProfile directly saves profile data to repository
func (s *ProfileServiceImpl) SaveProfile(ctx context.Context, profile models.Profile) error {
	return s.repo.SaveProfile(ctx, profile)
}
