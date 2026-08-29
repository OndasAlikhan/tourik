package domain

import (
	"errors"
	"time"
)

var (
	ErrParticipantNotFound = errors.New("participant not found")
)

type Participant struct {
	ID        int
	Name      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type ParticipantFilter struct {
	TournamentID int
}
