package tournament

import "time"

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
