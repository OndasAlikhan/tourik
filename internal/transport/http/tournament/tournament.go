package tournament

import (
	"context"
	"net/http"

	"github.com/OndasAlikhan/tourik/internal/domain"
	tournamentService "github.com/OndasAlikhan/tourik/internal/service/tournament"
	"github.com/gin-gonic/gin"
)

type Service interface {
	List(ctx context.Context) (result []domain.Tournament, err error)
}
type Handler struct {
	tournamentService Service
}

func New(tournamentService tournamentService.Service) Handler {
	return Handler{
		tournamentService: tournamentService,
	}
}

// ListTournaments godoc
//
//	@Summary		List tournaments
//	@Description	Returns a list of all tournaments
//	@Tags			tournaments
//	@Produce		json
//	@Success		200	{array}		Response
//	@Failure		500	{object}	nil
//	@Router			/tournaments [get]
func (h Handler) ListTournaments(ctx *gin.Context) {
	result, err := h.tournamentService.List(ctx)
	if err != nil {
		ctx.Status(http.StatusInternalServerError)
	}
	ctx.JSON(http.StatusOK, result)
}
