package tournament

import (
	"time"

	tourDomain "github.com/OndasAlikhan/tourik/internal/domain/tournament"
)

type Response struct {
	ID              int       `json:"id"`
	Name            string    `json:"name"`
	Discipline      string    `json:"discipline"`
	Status          string    `json:"status"`
	BracketFormat   string    `json:"bracketFormat"`
	MaxParticipants int       `json:"maxParticipants"`
	StartDate       time.Time `json:"startDate"`
	CreatedAt       time.Time `json:"createdAt"`
	UpdatedAt       time.Time `json:"updatedAt"`
}

type CreateTournamentRequest struct {
	Name            string    `json:"name"`
	Discipline      string    `json:"discipline"`
	Status          string    `json:"status"`
	BracketFormat   string    `json:"bracketFormat"`
	MaxParticipants int       `json:"maxParticipants"`
	StartDate       time.Time `json:"startDate"`
	EndDate         time.Time `json:"endDate"`
}

func (c CreateTournamentRequest) ToDomain() tourDomain.Tournament {
	return tourDomain.Tournament{
		Name:            c.Name,
		Discipline:      c.Discipline,
		Status:          tourDomain.TournamentStatus(c.Status),
		BracketFormat:   tourDomain.BracketFormat(c.BracketFormat),
		MaxParticipants: c.MaxParticipants,
		StartDate:       c.StartDate,
		EndDate:         c.EndDate,
	}
}
