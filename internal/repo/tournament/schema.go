package tournament

import (
	"time"

	tourDomain "github.com/OndasAlikhan/tourik/internal/domain/tournament"
)

type Tournament struct {
	ID              int       `db:"id"`
	Name            string    `db:"name"`
	Discipline      string    `db:"discipline"`
	Status          string    `db:"status"`
	BracketFormat   string    `db:"bracket_format"`
	MaxParticipants int       `db:"max_participants"`
	StartDate       time.Time `db:"start_date"`
	EndDate         time.Time `db:"end_date"`
	CreatedAt       time.Time `db:"created_at"`
	UpdatedAt       time.Time `db:"updated_at"`
}

func (t Tournament) ToDomain() tourDomain.Tournament {
	return tourDomain.Tournament{
		ID:              t.ID,
		Name:            t.Name,
		Discipline:      t.Discipline,
		Status:          tourDomain.TournamentStatus(t.Status),
		BracketFormat:   tourDomain.BracketFormat(t.BracketFormat),
		MaxParticipants: t.MaxParticipants,
		StartDate:       t.StartDate,
		EndDate:         t.EndDate,
		CreatedAt:       t.CreatedAt,
		UpdatedAt:       t.UpdatedAt,
	}
}

func FromDomain(t tourDomain.Tournament) Tournament {
	return Tournament{
		ID:              t.ID,
		Name:            t.Name,
		Discipline:      t.Discipline,
		Status:          string(t.Status),
		BracketFormat:   string(t.BracketFormat),
		MaxParticipants: t.MaxParticipants,
		StartDate:       t.StartDate,
		EndDate:         t.EndDate,
		CreatedAt:       t.CreatedAt,
		UpdatedAt:       t.UpdatedAt,
	}
}
