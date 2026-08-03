package domain

import "time"

type Tournament struct {
	ID              int
	Name            string
	Discipline      string
	BracketFormat   string
	MaxParticipants int
	StartDate       time.Time
	CreatedAt       time.Time
	UpdatedAt       time.Time
}
