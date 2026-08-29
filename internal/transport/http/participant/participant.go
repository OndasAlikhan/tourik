package participant

import (
	"context"
	"net/http"
	"strconv"

	"github.com/OndasAlikhan/tourik/internal/domain"
	"github.com/OndasAlikhan/tourik/internal/transport/http/wrap"
	participantUc "github.com/OndasAlikhan/tourik/internal/usecase/participant"
	"github.com/gin-gonic/gin"
)

type Usecase interface {
	List(ctx context.Context, filter domain.ParticipantFilter) (result []domain.Participant, err error)
	ByID(ctx context.Context, id int) (domain.Participant, error)
	Create(ctx context.Context, p domain.Participant) (domain.Participant, error)
	Update(ctx context.Context, p domain.Participant) (domain.Participant, error)
	CreateForTournament(ctx context.Context, tournamentID int, p domain.Participant) (domain.Participant, error)
	DeleteFromTournament(ctx context.Context, tournamentID, participantID int) error
}
type Handler struct {
	participantUc Usecase
}

func New(participantUsecase participantUc.Usecase) Handler {
	return Handler{
		participantUc: participantUsecase,
	}
}

// ListParticipants godoc
//
//	@Summary		List participants
//	@Description	Returns a list of all participants
//	@Tags			participants
//	@Produce		json
//	@Param			tournamentId	query		int	false	"Filter by tournament ID"
//	@Success		200	{array}		Response
//	@Failure		500	{object}	nil
//	@Router			/participants [get]
func (h Handler) ListParticipants(ctx *gin.Context) {
	var filter domain.ParticipantFilter
	if raw := ctx.Query("tournamentId"); raw != "" {
		tournamentID, err := strconv.Atoi(raw)
		if err != nil {
			wrap.WrapError(ctx, wrap.ErrMapObj{
				StatusCode: http.StatusBadRequest,
				Message:    "invalid tournamentId",
			})
			return
		}
		filter.TournamentID = tournamentID
	}

	result, err := h.participantUc.List(ctx, filter)
	if err != nil {
		wrap.WrapError(ctx, ErrorMap(err))
		return
	}
	wrap.Wrap(ctx, http.StatusOK, result)
}

// GetParticipant godoc
//
//	@Summary		Get participant
//	@Description	Returns a participant by id
//	@Tags			participants
//	@Produce		json
//	@Param			id	path		int	true	"Participant ID"
//	@Success		200	{object}	Response
//	@Failure		400	{object}	nil
//	@Failure		404	{object}	nil
//	@Router			/participants/{id} [get]
func (h Handler) GetParticipant(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		wrap.WrapError(ctx, wrap.ErrMapObj{
			StatusCode: http.StatusBadRequest,
			Message:    "invalid participant id",
		})
		return
	}

	result, err := h.participantUc.ByID(ctx, id)
	if err != nil {
		wrap.WrapError(ctx, ErrorMap(err))
		return
	}
	wrap.Wrap(ctx, http.StatusOK, result)
}

// CreateParticipant godoc
//
//	@Summary		Create participant
//	@Description	Creates a new participant
//	@Tags			participants
//	@Accept			json
//	@Produce		json
//	@Param			participant	body		CreateParticipant	true	"Participant"
//	@Success		201			{object}	Response
//	@Failure		400			{object}	nil
//	@Router			/participants [post]
func (h Handler) CreateParticipant(ctx *gin.Context) {
	var p CreateParticipant

	if err := ctx.ShouldBind(&p); err != nil {
		wrap.WrapError(ctx, wrap.ErrMapObj{
			StatusCode: http.StatusBadRequest,
			Message:    err.Error(),
		})
		return
	}

	result, err := h.participantUc.Create(ctx, p.ToDomain())
	if err != nil {
		wrap.WrapError(ctx, ErrorMap(err))
		return
	}
	wrap.Wrap(ctx, http.StatusCreated, result)
}

// UpdateParticipant godoc
//
//	@Summary		Update participant
//	@Description	Updates an existing participant
//	@Tags			participants
//	@Accept			json
//	@Produce		json
//	@Param			id			path		int					true	"Participant ID"
//	@Param			participant	body		CreateParticipant	true	"Participant"
//	@Success		200			{object}	Response
//	@Failure		400			{object}	nil
//	@Failure		404			{object}	nil
//	@Router			/participants/{id} [put]
func (h Handler) UpdateParticipant(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		wrap.WrapError(ctx, wrap.ErrMapObj{
			StatusCode: http.StatusBadRequest,
			Message:    "invalid participant id",
		})
		return
	}

	var p CreateParticipant
	if err := ctx.ShouldBind(&p); err != nil {
		wrap.WrapError(ctx, wrap.ErrMapObj{
			StatusCode: http.StatusBadRequest,
			Message:    err.Error(),
		})
		return
	}

	domainParticipant := p.ToDomain()
	domainParticipant.ID = id

	result, err := h.participantUc.Update(ctx, domainParticipant)
	if err != nil {
		wrap.WrapError(ctx, ErrorMap(err))
		return
	}
	wrap.Wrap(ctx, http.StatusOK, result)
}

// CreateTournamentParticipant godoc
//
//	@Summary		Add participant to tournament
//	@Description	Creates a new participant and ties it to the tournament
//	@Tags			participants
//	@Accept			json
//	@Produce		json
//	@Param			id			path		int					true	"Tournament ID"
//	@Param			participant	body		CreateParticipant	true	"Participant"
//	@Success		201			{object}	Response
//	@Failure		400			{object}	nil
//	@Router			/tournaments/{id}/participants [post]
func (h Handler) CreateTournamentParticipant(ctx *gin.Context) {
	tournamentID, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		wrap.WrapError(ctx, wrap.ErrMapObj{
			StatusCode: http.StatusBadRequest,
			Message:    "invalid tournament id",
		})
		return
	}

	var p CreateParticipant
	if err := ctx.ShouldBind(&p); err != nil {
		wrap.WrapError(ctx, wrap.ErrMapObj{
			StatusCode: http.StatusBadRequest,
			Message:    err.Error(),
		})
		return
	}

	result, err := h.participantUc.CreateForTournament(ctx, tournamentID, p.ToDomain())
	if err != nil {
		wrap.WrapError(ctx, ErrorMap(err))
		return
	}
	wrap.Wrap(ctx, http.StatusCreated, result)
}

// DeleteTournamentParticipant godoc
//
//	@Summary		Remove participant from tournament
//	@Description	Unties the participant from the tournament and deletes it
//	@Tags			participants
//	@Produce		json
//	@Param			id	path	int	true	"Tournament ID"
//	@Param			pid	path	int	true	"Participant ID"
//	@Success		204	{object}	nil
//	@Failure		400	{object}	nil
//	@Failure		404	{object}	nil
//	@Router			/tournaments/{id}/participants/{pid} [delete]
func (h Handler) DeleteTournamentParticipant(ctx *gin.Context) {
	tournamentID, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		wrap.WrapError(ctx, wrap.ErrMapObj{
			StatusCode: http.StatusBadRequest,
			Message:    "invalid tournament id",
		})
		return
	}

	participantID, err := strconv.Atoi(ctx.Param("pid"))
	if err != nil {
		wrap.WrapError(ctx, wrap.ErrMapObj{
			StatusCode: http.StatusBadRequest,
			Message:    "invalid participant id",
		})
		return
	}

	if err := h.participantUc.DeleteFromTournament(ctx, tournamentID, participantID); err != nil {
		wrap.WrapError(ctx, ErrorMap(err))
		return
	}
	ctx.Status(http.StatusNoContent)
}
