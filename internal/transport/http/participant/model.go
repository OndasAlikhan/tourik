package participant

import (
	"time"

	"github.com/OndasAlikhan/tourik/internal/domain"
)

type Response struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type CreateParticipant struct {
	Name string `json:"name"`
}

func (c CreateParticipant) ToDomain() domain.Participant {
	return domain.Participant{
		Name: c.Name,
	}
}
