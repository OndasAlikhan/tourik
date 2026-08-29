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
		tournaments := api.Group("/tournaments")
		{
			tournaments.GET("", cnt.Handlers.TournamentHandler.ListTournaments)
			tournaments.GET("/:id", cnt.Handlers.TournamentHandler.GetTournament)
			tournaments.POST("", cnt.Handlers.TournamentHandler.CreateTournament)
			tournaments.PUT("/:id", cnt.Handlers.TournamentHandler.UpdateTournament)

			tournaments.POST("/:id/participants", cnt.Handlers.ParticipantHandler.CreateTournamentParticipant)
			tournaments.DELETE("/:id/participants/:pid", cnt.Handlers.ParticipantHandler.DeleteTournamentParticipant)

			tournaments.POST("/:id/registration", cnt.Handlers.TournamentHandler.BeginRegistration)
			tournaments.POST("/:id/start", cnt.Handlers.TournamentHandler.Start)
			tournaments.POST("/:id/cancel", cnt.Handlers.TournamentHandler.Cancel)
		}

		participants := api.Group("/participants")
		{
			participants.GET("", cnt.Handlers.ParticipantHandler.ListParticipants)
			participants.GET("/:id", cnt.Handlers.ParticipantHandler.GetParticipant)
			participants.POST("", cnt.Handlers.ParticipantHandler.CreateParticipant)
			participants.PUT("/:id", cnt.Handlers.ParticipantHandler.UpdateParticipant)
		}
	}

	return router
}
