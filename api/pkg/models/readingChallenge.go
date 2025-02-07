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
	Current    int     `json:"current"`
	Percentage float64 `json:"percentage"`
	Rate       struct {
		Current  float64 `json:"current"`
		Required float64 `json:"required"`
		Unit     string  `json:"unit"`
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
