package internal

import (
	tournamentHandler "github.com/OndasAlikhan/tourik/internal/transport/http/tournament"
)

type Handlers struct {
	TournamentHandler tournamentHandler.Handler
}
type Container struct {
	Handlers Handlers
}
