package controllers

import (
	"context"
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
	var user models.ParseUserRegister
	if err := c.BindJSON(&user); err != nil {
		c.String(http.StatusBadRequest, err.Error())
		return
	}

	token, err := uc.service.RegisterUser(context.Background(), &user)
	if err != nil {
		c.String(statusFromError(err), err.Error())
	}

	c.Writer.Header().Add("Authorization", fmt.Sprintf("Bearer %s", token))
	c.Status(http.StatusOK)
}

func (uc *UserController) Login(c *gin.Context) {
	var user models.ParseUserRegister
	if err := c.BindJSON(&user); err != nil {
		c.String(http.StatusBadRequest, err.Error())
		return
	}

	token, err := uc.service.LoginUser(context.Background(), &user)
	if err != nil {
		c.String(statusFromError(err), err.Error())
	}

	c.Writer.Header().Add("Authorization", fmt.Sprintf("Bearer %s", token))
	c.Status(http.StatusOK)
}

func (uc *UserController) GetBalance(c *gin.Context) {
	login := c.GetString("login")

	balance, err := uc.service.GetBalance(context.Background(), &models.User{
		Login: login,
	})
	if err != nil {
		c.String(statusFromError(err), err.Error())
	}

	c.JSON(http.StatusOK, balance)
}
