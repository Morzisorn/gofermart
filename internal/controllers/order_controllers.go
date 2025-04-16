package controllers

import (
	"context"
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
	login := c.GetString("login")

	number, err := c.GetRawData()
	if err != nil {
		c.String(http.StatusBadRequest, err.Error())
		return
	}

	err = oc.service.UploadOrder(context.Background(), login, string(number))
	if err != nil {
		c.String(statusFromError(err), err.Error())
	}
	c.Status(http.StatusAccepted)
}

func (oc *OrderController) GetUserOrders(c *gin.Context) {
	login := c.GetString("login")

	ord, err := oc.service.GetUserOrders(context.Background(), login)
	if err != nil {
		c.String(statusFromError(err), err.Error())
	}

	c.JSON(http.StatusOK, ord)
}

func (oc *OrderController) GetUserWithdrawals(c *gin.Context) {
	login := c.GetString("login")

	withdrawals, err := oc.service.GetUserWithdrawals(context.Background(), login)
	if err != nil {
		c.String(statusFromError(err), err.Error())
	}

	c.JSON(http.StatusOK, withdrawals)
}

func (oc *OrderController) Withdraw(c *gin.Context) {
	login := c.GetString("login")

	var w models.Withdrawal
	if err := c.BindJSON(&w); err != nil {
		c.String(http.StatusBadRequest, err.Error())
		return
	}

	err := oc.service.Withdraw(context.Background(), login, &w)
	if err != nil {
		c.String(statusFromError(err), err.Error())
	}

	c.Status(http.StatusOK)
}
