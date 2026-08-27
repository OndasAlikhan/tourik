package tournament

import (
	"errors"

	"github.com/OndasAlikhan/tourik/internal/domain"
	"github.com/OndasAlikhan/tourik/internal/transport/http/wrap"
)

var errorMap = map[error]wrap.ErrMapObj{
	domain.ErrTourNotFound: {
		StatusCode: 404,
		Message:    "Tournament not found",
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
