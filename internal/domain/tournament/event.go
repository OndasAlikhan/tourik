package tournament

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type TournamentEvent string

const (
	TournamentCreated           TournamentEvent = "TournamentCreated"
	ParticipantRegistered       TournamentEvent = "ParticipantRegistered"
	ParticipantWithdrawn        TournamentEvent = "ParticipantWithdrawn"
	TournamentRegistrationBegan TournamentEvent = "TournamentRegistrationBegan"
	TournamentStarted           TournamentEvent = "TournamentStarted"
	TournamentCancelled         TournamentEvent = "TournamentCancelled"
)

type TournamentEventMessage struct {
	EventID   uuid.UUID       `json:"event_id"`
	EventType TournamentEvent `json:"event_type"`
	OccuredAt time.Time       `json:"occured_at"`
	Payload   json.RawMessage `json:"payload"`
}
