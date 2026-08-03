package tournament

import (
	"net/http"

	tournamentService "github.com/OndasAlikhan/tourik/internal/service/tournament"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	tournamentService tournamentService.Service
}

func New(tournamentService tournamentService.Service) Handler {
	return Handler{
		tournamentService: tournamentService,
	}
}

func (h Handler) ListTournaments(ctx *gin.Context) {
	result, err := h.tournamentService.List(ctx)
	if err != nil {
		ctx.Status(http.StatusInternalServerError)
	}
	ctx.JSON(http.StatusOK, result)
}
