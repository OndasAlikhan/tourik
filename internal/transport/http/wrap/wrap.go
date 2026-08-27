package wrap

import "github.com/gin-gonic/gin"

type ErrMapObj struct {
	StatusCode int
	Message    string
}

type Response struct {
	Error string `json:"error_message"`
	Data  any    `json:"data"`
}

func Wrap(ctx *gin.Context, code int, data any) {
	ctx.JSON(code, Response{Data: data})
}

func WrapError(ctx *gin.Context, emObj ErrMapObj) {
	ctx.JSON(emObj.StatusCode, Response{Error: emObj.Message})
}
