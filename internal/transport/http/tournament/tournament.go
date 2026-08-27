package tournament

import (
	"context"
	"net/http"
	"strconv"

	"github.com/OndasAlikhan/tourik/internal/domain"
	"github.com/OndasAlikhan/tourik/internal/transport/http/wrap"
	tourUc "github.com/OndasAlikhan/tourik/internal/usecase/tournament"
	"github.com/gin-gonic/gin"
)

type Usecase interface {
	List(ctx context.Context) (result []domain.Tournament, err error)
	ByID(ctx context.Context, id int) (domain.Tournament, error)
	Create(ctx context.Context, t domain.Tournament) (domain.Tournament, error)
	Update(ctx context.Context, t domain.Tournament) (domain.Tournament, error)
}
type Handler struct {
	tourUc Usecase
}

func New(tournamentUsecase tourUc.Usecase) Handler {
	return Handler{
		tourUc: tournamentUsecase,
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
	result, err := h.tourUc.List(ctx)
	if err != nil {
		wrap.WrapError(ctx, ErrorMap(err))
		return
	}
	wrap.Wrap(ctx, http.StatusOK, result)
}

// GetTournament godoc
//
//	@Summary		Get tournament
//	@Description	Returns a tournament by id
//	@Tags			tournaments
//	@Produce		json
//	@Param			id	path		int	true	"Tournament ID"
//	@Success		200	{object}	Response
//	@Failure		400	{object}	nil
//	@Failure		404	{object}	nil
//	@Router			/tournaments/{id} [get]
func (h Handler) GetTournament(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		wrap.WrapError(ctx, wrap.ErrMapObj{
			StatusCode: http.StatusBadRequest,
			Message:    "invalid tournament id",
		})
		return
	}

	result, err := h.tourUc.ByID(ctx, id)
	if err != nil {
		wrap.WrapError(ctx, ErrorMap(err))
		return
	}
	wrap.Wrap(ctx, http.StatusOK, result)
}

// CreateTournament godoc
//
//	@Summary		Create tournament
//	@Description	Creates a new tournament
//	@Tags			tournaments
//	@Accept			json
//	@Produce		json
//	@Param			tournament	body		CreateTournament	true	"Tournament"
//	@Success		201			{object}	Response
//	@Failure		400			{object}	nil
//	@Router			/tournaments [post]
func (h Handler) CreateTournament(ctx *gin.Context) {
	var tour CreateTournament

	if err := ctx.ShouldBind(&tour); err != nil {
		wrap.WrapError(ctx, wrap.ErrMapObj{
			StatusCode: http.StatusBadRequest,
			Message:    err.Error(),
		})
		return
	}

	result, err := h.tourUc.Create(ctx, tour.ToDomain())
	if err != nil {
		wrap.WrapError(ctx, ErrorMap(err))
		return
	}
	wrap.Wrap(ctx, http.StatusCreated, result)
}

// UpdateTournament godoc
//
//	@Summary		Update tournament
//	@Description	Updates an existing tournament
//	@Tags			tournaments
//	@Accept			json
//	@Produce		json
//	@Param			id			path		int					true	"Tournament ID"
//	@Param			tournament	body		CreateTournament	true	"Tournament"
//	@Success		200			{object}	Response
//	@Failure		400			{object}	nil
//	@Failure		404			{object}	nil
//	@Router			/tournaments/{id} [put]
func (h Handler) UpdateTournament(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		wrap.WrapError(ctx, wrap.ErrMapObj{
			StatusCode: http.StatusBadRequest,
			Message:    "invalid tournament id",
		})
		return
	}

	var tour CreateTournament
	if err := ctx.ShouldBind(&tour); err != nil {
		wrap.WrapError(ctx, wrap.ErrMapObj{
			StatusCode: http.StatusBadRequest,
			Message:    err.Error(),
		})
		return
	}

	domainTour := tour.ToDomain()
	domainTour.ID = id

	result, err := h.tourUc.Update(ctx, domainTour)
	if err != nil {
		wrap.WrapError(ctx, ErrorMap(err))
		return
	}
	wrap.Wrap(ctx, http.StatusOK, result)
}
