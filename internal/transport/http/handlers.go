package http

import (
	"github.com/OndasAlikhan/tourik/internal"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func Routers(cnt internal.Container) *gin.Engine {
	router := gin.Default()

	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	api := router.Group("/api")
	{
		api.GET("/tournaments", cnt.Handlers.TournamentHandler.ListTournaments)
		api.GET("/tournaments/:id", cnt.Handlers.TournamentHandler.GetTournament)
		api.POST("/tournaments", cnt.Handlers.TournamentHandler.CreateTournament)
		api.PUT("/tournaments/:id", cnt.Handlers.TournamentHandler.UpdateTournament)
	}

	return router
}
