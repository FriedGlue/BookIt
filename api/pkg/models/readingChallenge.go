package models

import (
	"time"
)

type ChallengeType string
type TimeFrame string

const (
	BooksChallenge ChallengeType = "BOOKS"
	PagesChallenge ChallengeType = "PAGES"

	YearTimeFrame  TimeFrame = "YEAR"
	MonthTimeFrame TimeFrame = "MONTH"
	WeekTimeFrame  TimeFrame = "WEEK"
)

type ChallengeProgress struct {
	Current    int     `json:"current"`    // Total progress (pages or books read)
	Percentage float64 `json:"percentage"` // Completion percentage (0-100)
	Rate       struct {
		Required     float64 `json:"required"`     // Target pace needed (pages/day, books/day, etc.)
		CurrentPace  float64 `json:"currentPace"`  // Actual reading pace (pages/day, books/day, etc.)
		ScheduleDiff float64 `json:"scheduleDiff"` // Cumulative difference between expected and actual progress
		Unit         string  `json:"unit"`         // Unit for the pace (pages/day, books/month, etc.)
		Status       string  `json:"status"`       // "AHEAD", "BEHIND", or "ON_TRACK"
	} `json:"rate"`
}

type ReadingChallenge struct {
	ID        string            `json:"id" dynamodbav:"id"`
	UserID    string            `json:"userId" dynamodbav:"userId"`
	Name      string            `json:"name" dynamodbav:"name"`
	Type      ChallengeType     `json:"type" dynamodbav:"type"`
	TimeFrame TimeFrame         `json:"timeframe" dynamodbav:"timeframe"`
	StartDate time.Time         `json:"startDate" dynamodbav:"startDate"`
	EndDate   time.Time         `json:"endDate" dynamodbav:"endDate"`
	Target    int               `json:"target" dynamodbav:"target"`
	Progress  ChallengeProgress `json:"progress" dynamodbav:"progress"`
	CreatedAt time.Time         `json:"createdAt" dynamodbav:"createdAt"`
	UpdatedAt time.Time         `json:"updatedAt" dynamodbav:"updatedAt"`
}
