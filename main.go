package main

import (
	"context"
	"net/http"

	"github.com/crusyn/loans/ent"
	"github.com/crusyn/loans/handlers"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"

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
	r.GET("/", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{
			"message": "server is up",
		})
	})

	r.POST("/user", h.CreateUser)
	r.POST("/loan", h.CreateLoan)
	r.GET("/loan/:id", h.GetLoan)
	r.GET("/loan/:id/schedule", h.GetLoanSchedule)
	r.GET("/loan/:id/month/:number/", h.GetMonthSummary)

	r.Run() // listen and serve on 0.0.0.0:8080
}

// func main() {
// 	s, _ := handlers.CreateAmortizationSchedule(1212530.0, 0.115, 360)

// 	for _, r := range s {
// 		fmt.Printf("r: %v\n", r)
// 	}
// }
