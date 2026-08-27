package tournament

import (
	"time"

	"github.com/OndasAlikhan/tourik/internal/domain"
)

type Response struct {
	ID              int       `json:"id"`
	Name            string    `json:"name"`
	Discipline      string    `json:"discipline"`
	BracketFormat   string    `json:"bracketFormat"`
	MaxParticipants int       `json:"maxParticipants"`
	StartDate       time.Time `json:"startDate"`
	CreatedAt       time.Time `json:"createdAt"`
	UpdatedAt       time.Time `json:"updatedAt"`
}

type CreateTournament struct {
	Name            string    `json:"name"`
	Discipline      string    `json:"discipline"`
	BracketFormat   string    `json:"bracketFormat"`
	MaxParticipants int       `json:"maxParticipants"`
	StartDate       time.Time `json:"startDate"`
	EndDate         time.Time `json:"endDate"`
}

func (c CreateTournament) ToDomain() domain.Tournament {
	return domain.Tournament{
		Name:            c.Name,
		Discipline:      c.Discipline,
		BracketFormat:   c.BracketFormat,
		MaxParticipants: c.MaxParticipants,
		StartDate:       c.StartDate,
		EndDate:         c.EndDate,
	}
}
