// pkg/usecase/challenge.go
package usecase

import (
	"context"

	"github.com/FriedGlue/BookIt/api/pkg/models"
)

// ReadingChallengeService defines operations for reading challenges
type ReadingChallengeService interface {
	// Challenge management
	GetChallenges(ctx context.Context, userID string) ([]models.ReadingChallenge, error)
	CreateChallenge(ctx context.Context, userID string, challenge models.ReadingChallenge) error
	UpdateChallenge(ctx context.Context, userID string, challengeID string, updates map[string]interface{}) error
	DeleteChallenge(ctx context.Context, userID string, challengeID string) error

	// Challenge progress
	UpdateChallengeProgress(ctx context.Context, userID string) error
}
