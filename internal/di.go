package internal

import (
	participantHandler "github.com/OndasAlikhan/tourik/internal/transport/http/participant"
	tournamentHandler "github.com/OndasAlikhan/tourik/internal/transport/http/tournament"
)

type Handlers struct {
	TournamentHandler  tournamentHandler.Handler
	ParticipantHandler participantHandler.Handler
}
type Container struct {
	Handlers Handlers
}
