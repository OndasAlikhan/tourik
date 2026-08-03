package http

import (
	"github.com/OndasAlikhan/tourik/internal"
	"github.com/gin-gonic/gin"
)

func Routers(cnt internal.Container) *gin.Engine {
	router := gin.Default()
	api := router.Group("/api")
	{
		api.GET("/tournaments", cnt.Handlers.TournamentHandler.ListTournaments)
	}

	return router
}
