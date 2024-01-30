package main

import (
	"context"
	"net/http"

	"github.com/crusyn/loans/docs"
	"github.com/crusyn/loans/ent"
	"github.com/crusyn/loans/handlers"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"

	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	// db init
	client, err := ent.Open("sqlite3", "file:ent?mode=memory&cache=shared&_fk=1")
	if err != nil {
		log.Fatal().Msgf("failed opening connection to sqlite: %v", err)
	}
	defer client.Close()
	// Run the auto migration tool.
	if err := client.Schema.Create(context.Background()); err != nil {
		log.Fatal().Msgf("failed creating schema resources: %v", err)
	}

	h := handlers.Handler{
		Ent: client,
	}

	// Server init
	r := gin.Default()
	docs.SwaggerInfo.BasePath = "/"
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))

	r.GET("/", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{
			"message": "server is up",
		})
	})

	r.POST("/user", h.CreateUser)
	r.GET("/user/:id/loans", h.GetLoans)
	r.POST("/loan", h.CreateLoan)
	r.GET("/loan/:id", h.GetLoan)
	r.GET("/loan/:id/schedule", h.GetLoanSchedule)
	r.GET("/loan/:id/month/:number/", h.GetMonthSummary)
	r.POST("loan/:id/share", h.ShareLoan)

	r.Run() // listen and serve on 0.0.0.0:8080
}
