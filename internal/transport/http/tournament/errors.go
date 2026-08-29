package tournament

import (
	"errors"

	tourDomain "github.com/OndasAlikhan/tourik/internal/domain/tournament"
	"github.com/OndasAlikhan/tourik/internal/transport/http/wrap"
)

var errorMap = map[error]wrap.ErrMapObj{
	tourDomain.ErrTourNotFound: {
		StatusCode: 404,
		Message:    "Tournament not found",
	},
	tourDomain.ErrWrongStatus: {
		StatusCode: 409,
		Message:    "Tournament status does not allow this action",
	},
	tourDomain.ErrMaxParticipants: {
		StatusCode: 409,
		Message:    "Tournament does not meet participant requirements",
	},
}

func ErrorMap(err error) wrap.ErrMapObj {
	for sentinel, eObj := range errorMap {
		if errors.Is(err, sentinel) {
			return eObj
		}
	}

	return wrap.ErrMapObj{
		StatusCode: 500,
		Message:    "Server error",
	}
}
