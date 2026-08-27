package domain

import (
	"errors"
	"time"
)

var (
	ErrTourNotFound = errors.New("tournament not found")
)

type Tournament struct {
	ID              int
	Name            string
	Discipline      string
	BracketFormat   string
	MaxParticipants int
	StartDate       time.Time
	EndDate         time.Time
	CreatedAt       time.Time
	UpdatedAt       time.Time
}
