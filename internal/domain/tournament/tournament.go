package tournament

import (
	"errors"
	"fmt"
	"slices"
	"time"

	"github.com/OndasAlikhan/tourik/internal/domain"
)

var (
	ErrTourNotFound    = errors.New("tournament not found")
	ErrWrongStatus     = errors.New("wrong status")
	ErrMaxParticipants = errors.New("wrong number of participants")
)

type Tournament struct {
	ID              int
	Name            string
	Discipline      string
	Status          TournamentStatus
	BracketFormat   BracketFormat
	MaxParticipants int
	StartDate       time.Time
	EndDate         time.Time
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type TournamentStatus string

const (
	Draft        TournamentStatus = "draft"
	Registration TournamentStatus = "registration"
	InProgress   TournamentStatus = "in_progress"
	Finished     TournamentStatus = "finished"
	Cancelled    TournamentStatus = "cancelled"
)

type BracketFormat string

const (
	SingleElimination BracketFormat = "SingleElimination"
	DoubleElimination BracketFormat = "DoubleElimination"
	RoundRobin        BracketFormat = "RoundRobin"
)

func (t *Tournament) CheckConditions(pts []domain.Participant) error {
	if len(pts) < 2 {
		return fmt.Errorf("%w: required more than one participant", ErrMaxParticipants)
	}
	if slices.Contains([]BracketFormat{SingleElimination, DoubleElimination}, t.BracketFormat) && !isPowerOfTwo(t.MaxParticipants) {
		return fmt.Errorf("%w: number of participants must be power of two for elimination", ErrMaxParticipants)
	}
	return nil
}

func (t *Tournament) BeginRegistration() error {
	if t.Status != Draft {
		return fmt.Errorf("cannot begin registration, status must be %s: %w", Draft, ErrWrongStatus)
	}
	t.Status = Registration

	return nil
}

func (t *Tournament) Start() error {
	if t.Status != Registration {
		return fmt.Errorf("cannot begin registration, status must be %s: %w", Registration, ErrWrongStatus)
	}
	t.Status = InProgress

	return nil
}

func (t *Tournament) Cancel() error {
	if t.Status == Finished || t.Status == Cancelled {
		return fmt.Errorf("cannot cancel tournament with status %s: %w", t.Status, ErrWrongStatus)
	}
	t.Status = Cancelled

	return nil
}

func isPowerOfTwo(n int) bool {
	return n > 0 && n&(n-1) == 0
}
