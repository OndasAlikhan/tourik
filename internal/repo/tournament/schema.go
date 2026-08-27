package tournament

import (
	"time"

	"github.com/OndasAlikhan/tourik/internal/domain"
)

type Tournament struct {
	ID              int       `db:"id"`
	Name            string    `db:"name"`
	Discipline      string    `db:"discipline"`
	BracketFormat   string    `db:"bracket_format"`
	MaxParticipants int       `db:"max_participants"`
	StartDate       time.Time `db:"start_date"`
	EndDate         time.Time `db:"end_date"`
	CreatedAt       time.Time `db:"created_at"`
	UpdatedAt       time.Time `db:"updated_at"`
}

func (t Tournament) ToDomain() domain.Tournament {
	return domain.Tournament{
		ID:              t.ID,
		Name:            t.Name,
		Discipline:      t.Discipline,
		BracketFormat:   t.BracketFormat,
		MaxParticipants: t.MaxParticipants,
		StartDate:       t.StartDate,
		EndDate:         t.EndDate,
		CreatedAt:       t.CreatedAt,
		UpdatedAt:       t.UpdatedAt,
	}
}

func FromDomain(t domain.Tournament) Tournament {
	return Tournament{
		ID:              t.ID,
		Name:            t.Name,
		Discipline:      t.Discipline,
		BracketFormat:   t.BracketFormat,
		MaxParticipants: t.MaxParticipants,
		StartDate:       t.StartDate,
		EndDate:         t.EndDate,
		CreatedAt:       t.CreatedAt,
		UpdatedAt:       t.UpdatedAt,
	}
}
