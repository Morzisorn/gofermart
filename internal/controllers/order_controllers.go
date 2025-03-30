package controllers

import (
	"context"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/morzisorn/gofermart/internal/models"
	"github.com/morzisorn/gofermart/internal/services/orders"
)

type OrderController struct {
	service *orders.OrderService
}

func NewOrderController(s *orders.OrderService) *OrderController {
	return &OrderController{service: s}
}

func (oc *OrderController) UploadOrder(c *gin.Context) {
	if c.Request.Method != http.MethodPost {
		c.String(http.StatusBadRequest, "Invalid request method")
		return
	}

	if c.Request.Header.Get("Content-Type") != "text/plain" {
		c.String(http.StatusBadRequest, "Invalid content type")
		return
	}

	login := c.GetString("login")

	number, err := c.GetRawData()
	if err != nil {
		c.String(http.StatusBadRequest, err.Error())
		return
	}

	err = oc.service.UploadOrder(context.Background(), login, string(number))
	switch {
	case errors.Is(err, orders.ErrIncorrectNumber):
		c.String(http.StatusUnprocessableEntity, err.Error())
		return
	case errors.Is(err, orders.ErrOrderAlreadyExist):
		c.String(http.StatusOK, err.Error())
	case errors.Is(err, orders.ErrOrderBelongsAnotherUser):
		c.String(http.StatusConflict, err.Error())
	case err != nil:
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	c.Status(http.StatusAccepted)
}

func (oc *OrderController) GetUserOrders(c *gin.Context) {
	if c.Request.Method != http.MethodGet {
		c.String(http.StatusBadRequest, "Invalid request method")
		return
	}

	login := c.GetString("login")

	ord, err := oc.service.GetUserOrders(context.Background(), login)
	switch {
	case errors.Is(err, orders.ErrNoData):
		c.Status(http.StatusNoContent)
		return
	case err != nil:
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, ord)
}

func (oc *OrderController) GetUserWithdrawals(c *gin.Context) {
	if c.Request.Method != http.MethodGet {
		c.String(http.StatusBadRequest, "Invalid request method")
		return
	}

	login := c.GetString("login")

	withdrawals, err := oc.service.GetUserWithdrawals(context.Background(), login)
	switch {
	case errors.Is(err, orders.ErrNoData):
		c.Status(http.StatusNoContent)
		return
	case err != nil:
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, withdrawals)
}

func (oc *OrderController) Withdraw(c *gin.Context) {
	if c.Request.Method != http.MethodPost {
		c.String(http.StatusBadRequest, "Invalid request method")
		return
	}

	login := c.GetString("login")

	var w models.Withdrawal
	if err := c.BindJSON(&w); err != nil {
		c.String(http.StatusBadRequest, err.Error())
		return
	}

	err := oc.service.Withdraw(context.Background(), login, &w)
	switch {
	case errors.Is(err, orders.ErrIncorrectNumber):
		c.String(http.StatusUnprocessableEntity, err.Error())
		return
	case errors.Is(err, orders.ErrInsufficientBalance):
		c.String(http.StatusPaymentRequired, err.Error())
		return
	case err != nil:
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	c.Status(http.StatusOK)
}
