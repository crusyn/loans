package main

import (
	"net/http"

	"github.com/crusyn/loans/ent"
	"github.com/crusyn/loans/ent/user"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

type Handler struct {
	Ent *ent.Client
}

type userBody struct {
	Name    string `json:"name"`
	Social  string `json:"social"`
	Address string `json:"address"`
}

func (h Handler) CreateUser(ctx *gin.Context) {

	var newUser userBody

	if err := ctx.BindJSON(&newUser); err != nil {
		log.Debug().Msgf("%v", err)
		ctx.JSON(http.StatusUnprocessableEntity, gin.H{
			"message": "user input malformed",
		})
		return
	}

	socialExists, err := h.Ent.User.Query().Where(user.Social(newUser.Social)).Exist(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"message": "internal error",
		})
		return
	}

	if socialExists {
		ctx.JSON(http.StatusUnprocessableEntity, gin.H{
			"message": "user with social security number already exists",
		})
		return
	}

	if err = h.Ent.User.Create().
		SetName(newUser.Name).
		SetSocial(newUser.Social).
		SetAddress(newUser.Address).
		Exec(ctx); err != nil {
		log.Err(err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"message": "internal error",
		})
		return
	}

}
