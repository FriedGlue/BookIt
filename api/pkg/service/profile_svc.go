// pkg/service/profile_svc.go
package service

import (
	"context"

	"github.com/FriedGlue/BookIt/api/pkg/models"
	"github.com/FriedGlue/BookIt/api/pkg/usecase"
)

type profileSvc struct {
	repo usecase.ProfileRepo
}

// NewProfileService constructs ProfileService.
func NewProfileService(repo usecase.ProfileRepo) usecase.ProfileService {
	return &profileSvc{
		repo: repo,
	}
}

// GetProfile loads a user's profile from repository
func (s *profileSvc) GetProfile(ctx context.Context, userID string) (models.Profile, error) {
	return s.repo.LoadProfile(ctx, userID)
}

// SaveProfile saves a user's profile to repository
func (s *profileSvc) SaveProfile(ctx context.Context, profile models.Profile) error {
	return s.repo.SaveProfile(ctx, profile)
}
