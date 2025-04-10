package controllers

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/morzisorn/gofermart/internal/models"
	"github.com/morzisorn/gofermart/internal/services/users"
)

type UserController struct {
	service *users.UserService
}

func NewUserController(s *users.UserService) *UserController {
	return &UserController{service: s}
}

func (uc *UserController) RegisterUser(c *gin.Context) {
	if c.Request.Method != http.MethodPost {
		c.String(http.StatusBadRequest, "Invalid request method")
		return
	}

	if c.Request.Header.Get("Content-Type") != "application/json" {
		c.String(http.StatusBadRequest, "Invalid content type")
		return
	}

	var user models.ParseUserRegister
	if err := c.BindJSON(&user); err != nil {
		c.String(http.StatusBadRequest, err.Error())
		return
	}

	token, err := uc.service.RegisterUser(context.Background(), &user)
	switch {
	case errors.Is(err, users.ErrUserAlreadyRegistered):
		c.String(http.StatusConflict, users.ErrUserAlreadyRegistered.Error())
		return
	case err != nil:
		c.String(http.StatusInternalServerError, "internal server error")
		return
	}

	c.Writer.Header().Add("Authorization", fmt.Sprintf("Bearer %s", token))
	c.Status(http.StatusOK)
}

func (uc *UserController) Login(c *gin.Context) {
	if c.Request.Method != http.MethodPost {
		c.String(http.StatusBadRequest, "Invalid request method")
		return
	}

	if c.Request.Header.Get("Content-Type") != "application/json" {
		c.String(http.StatusBadRequest, "Invalid content type")
		return
	}

	var user models.ParseUserRegister
	if err := c.BindJSON(&user); err != nil {
		c.String(http.StatusBadRequest, err.Error())
		return
	}

	token, err := uc.service.LoginUser(context.Background(), &user)
	switch {
	case errors.Is(err, users.ErrIncorrectCredentials):
		c.String(http.StatusUnauthorized, users.ErrIncorrectCredentials.Error())
		return
	case err != nil:
		c.String(http.StatusInternalServerError, users.ErrInternalServerError.Error())
		return
	}

	c.Writer.Header().Add("Authorization", fmt.Sprintf("Bearer %s", token))
	c.Status(http.StatusOK)
}

func (uc *UserController) GetBalance(c *gin.Context) {
	if c.Request.Method != http.MethodGet {
		c.String(http.StatusBadRequest, "Invalid request method")
		return
	}

	login := c.GetString("login")

	user := models.User{
		Login: login,
	}

	balance, err := uc.service.GetBalance(context.Background(), &user)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, balance)
}
