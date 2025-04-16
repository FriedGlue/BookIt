// pkg/service/challenge_svc.go
package service

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/FriedGlue/BookIt/api/pkg/models"
	"github.com/FriedGlue/BookIt/api/pkg/usecase"
)

type challengeSvc struct {
	repo       usecase.ProfileRepo
	readingSvc usecase.ReadingListService
}

// NewChallengeService creates a new reading challenge service
func NewChallengeService(r usecase.ProfileRepo, readingSvc usecase.ReadingListService) usecase.ReadingChallengeService {
	return &challengeSvc{
		repo:       r,
		readingSvc: readingSvc,
	}
}

func (s *challengeSvc) GetChallenges(ctx context.Context, userID string) ([]models.ReadingChallenge, error) {
	profile, err := s.repo.LoadProfile(ctx, userID)
	if err != nil {
		return nil, err
	}
	return profile.Challenges, nil
}

func (s *challengeSvc) CreateChallenge(ctx context.Context, userID string, challenge models.ReadingChallenge) error {
	profile, err := s.repo.LoadProfile(ctx, userID)
	if err != nil {
		return err
	}

	now := time.Now()

	// Set required fields
	challenge.ID = fmt.Sprintf("challenge_%d", now.UnixNano())
	challenge.UserID = userID
	challenge.CreatedAt = now
	challenge.UpdatedAt = now

	// Initialize progress
	challenge.Progress = models.ChallengeProgress{
		Current:    0,
		Percentage: 0,
	}

	// Set required reading rate
	requiredRate, unit := calculateRequiredRate(challenge)
	challenge.Progress.Rate.Required = requiredRate
	challenge.Progress.Rate.Unit = unit
	challenge.Progress.Rate.Status = "ON_TRACK"

	// Add challenge to profile
	profile.Challenges = append(profile.Challenges, challenge)

	return s.repo.SaveProfile(ctx, profile)
}

func (s *challengeSvc) UpdateChallenge(ctx context.Context, userID string, challengeID string, updates map[string]interface{}) error {
	profile, err := s.repo.LoadProfile(ctx, userID)
	if err != nil {
		return err
	}

	for i, challenge := range profile.Challenges {
		if challenge.ID == challengeID {
			// Apply updates
			if target, ok := updates["target"].(int); ok {
				profile.Challenges[i].Target = target
			}
			if name, ok := updates["name"].(string); ok {
				profile.Challenges[i].Name = name
			}
			if endDate, ok := updates["endDate"].(time.Time); ok {
				profile.Challenges[i].EndDate = endDate
			}
			if typ, ok := updates["type"].(string); ok {
				profile.Challenges[i].Type = models.ChallengeType(typ)
			}
			if timeFrame, ok := updates["timeframe"].(string); ok {
				profile.Challenges[i].TimeFrame = models.TimeFrame(timeFrame)
			}

			// Update timestamps and rates
			profile.Challenges[i].UpdatedAt = time.Now()
			requiredRate, unit := calculateRequiredRate(profile.Challenges[i])
			profile.Challenges[i].Progress.Rate.Required = requiredRate
			profile.Challenges[i].Progress.Rate.Unit = unit

			return s.repo.SaveProfile(ctx, profile)
		}
	}

	return fmt.Errorf("challenge not found: %s", challengeID)
}

func (s *challengeSvc) DeleteChallenge(ctx context.Context, userID string, challengeID string) error {
	profile, err := s.repo.LoadProfile(ctx, userID)
	if err != nil {
		return err
	}

	for i, challenge := range profile.Challenges {
		if challenge.ID == challengeID {
			// Remove challenge
			profile.Challenges = append(profile.Challenges[:i], profile.Challenges[i+1:]...)
			return s.repo.SaveProfile(ctx, profile)
		}
	}

	return fmt.Errorf("challenge not found: %s", challengeID)
}

func (s *challengeSvc) UpdateChallengeProgress(ctx context.Context, userID string) error {
	profile, err := s.repo.LoadProfile(ctx, userID)
	if err != nil {
		return err
	}

	// No challenges to update
	if len(profile.Challenges) == 0 {
		return nil
	}

	// Get reading data
	readItems, err := s.readingSvc.GetRead(ctx, userID)
	if err != nil {
		return err
	}

	updated := false
	now := time.Now()

	for i, challenge := range profile.Challenges {
		// Skip already completed challenges
		if challenge.Progress.Percentage >= 100 {
			continue
		}

		// Calculate progress based on challenge type
		progress := aggregateChallengeProgress(&profile, challenge, readItems)
		if progress != challenge.Progress.Current {
			profile.Challenges[i].Progress.Current = progress
			if challenge.Target > 0 {
				profile.Challenges[i].Progress.Percentage = math.Min(100, float64(progress)/float64(challenge.Target)*100)
			}

			// Update current pace
			profile.Challenges[i].Progress.Rate.CurrentPace = calculateCurrentPace(challenge, now, progress)

			// Update schedule difference and status
			scheduleDiff, status := calculateScheduleStatus(challenge, now, progress)
			profile.Challenges[i].Progress.Rate.ScheduleDiff = scheduleDiff
			profile.Challenges[i].Progress.Rate.Status = status

			profile.Challenges[i].UpdatedAt = now
			updated = true
		}
	}

	if updated {
		return s.repo.SaveProfile(ctx, profile)
	}

	return nil
}

// Helper functions

// calculateRequiredRate calculates the required reading pace based on challenge parameters
func calculateRequiredRate(challenge models.ReadingChallenge) (float64, string) {
	duration := challenge.EndDate.Sub(challenge.StartDate)
	var rate float64
	var unit string

	switch challenge.TimeFrame {
	case models.YearTimeFrame:
		monthsTotal := float64(duration.Hours()) / (24 * 30)
		rate = math.Round(float64(challenge.Target)/monthsTotal*100) / 100
		if challenge.Type == models.BooksChallenge {
			unit = "books/month"
		} else {
			unit = "pages/month"
		}
	case models.MonthTimeFrame:
		weeksTotal := float64(duration.Hours()) / (24 * 7)
		rate = math.Round(float64(challenge.Target)/weeksTotal*100) / 100
		if challenge.Type == models.BooksChallenge {
			unit = "books/week"
		} else {
			unit = "pages/week"
		}
	case models.WeekTimeFrame:
		daysTotal := float64(duration.Hours()) / 24
		rate = math.Round(float64(challenge.Target)/daysTotal*100) / 100
		if challenge.Type == models.BooksChallenge {
			unit = "books/day"
		} else {
			unit = "pages/day"
		}
	}

	return rate, unit
}

// calculateCurrentPace computes the actual reading pace
func calculateCurrentPace(challenge models.ReadingChallenge, now time.Time, progress int) float64 {
	duration := now.Sub(challenge.StartDate)
	if duration <= 0 {
		return 0
	}

	var divisor float64
	switch challenge.TimeFrame {
	case models.YearTimeFrame:
		divisor = float64(duration.Hours()) / (24 * 30) // months
	case models.MonthTimeFrame:
		divisor = float64(duration.Hours()) / (24 * 7) // weeks
	case models.WeekTimeFrame:
		divisor = float64(duration.Hours()) / 24 // days
	}

	if divisor <= 0 {
		return 0
	}

	return math.Round(float64(progress)/divisor*100) / 100
}

// calculateScheduleStatus determines if ahead/behind and by how much
func calculateScheduleStatus(challenge models.ReadingChallenge, now time.Time, progress int) (float64, string) {
	// If the challenge hasn't started yet, consider it on track
	if now.Before(challenge.StartDate) {
		return 0, "ON_TRACK"
	}

	duration := now.Sub(challenge.StartDate)
	totalDuration := challenge.EndDate.Sub(challenge.StartDate)

	if totalDuration <= 0 {
		return 0, "ON_TRACK"
	}

	// Calculate expected progress at this point
	expectedProgress := float64(challenge.Target) * (float64(duration) / float64(totalDuration))
	actualProgress := float64(progress)

	// Calculate the progress difference
	progressDiff := actualProgress - expectedProgress

	// If the difference is negligible, consider it on track
	if math.Abs(progressDiff) <= 0.15 {
		return 0, "ON_TRACK"
	}

	// Return the absolute difference and status
	if progressDiff > 0 {
		return math.Round(progressDiff*100) / 100, "AHEAD"
	}
	return math.Round(math.Abs(progressDiff)*100) / 100, "BEHIND"
}

// aggregateChallengeProgress counts progress for a challenge
func aggregateChallengeProgress(profile *models.Profile, challenge models.ReadingChallenge, readItems []models.ReadItem) int {
	total := 0
	switch challenge.Type {
	case models.BooksChallenge:
		// For books challenge, count books completed in the period
		for _, book := range readItems {
			if book.CompletedDate != "" {
				completedDate, err := time.Parse(time.RFC3339, book.CompletedDate)
				if err == nil &&
					!completedDate.Before(challenge.StartDate) &&
					(completedDate.Before(challenge.EndDate) || challenge.EndDate.IsZero()) {
					total++
				}
			}
		}
	case models.PagesChallenge:
		// For pages challenge, sum the pages from reading log
		for _, logEntry := range profile.ReadingLog {
			if logEntry.Date != "" {
				logDate, err := time.Parse(time.RFC3339, logEntry.Date)
				if err == nil &&
					!logDate.Before(challenge.StartDate) &&
					(logDate.Before(challenge.EndDate) || challenge.EndDate.IsZero()) {
					total += logEntry.PagesRead
				}
			}
		}
	}

	return total
}
