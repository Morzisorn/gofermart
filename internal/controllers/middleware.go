package controllers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func RequireContentType(expected string) gin.HandlerFunc {
	return func(c *gin.Context) {
		contentType := c.GetHeader("Content-Type")
		if !strings.HasPrefix(contentType, expected) {
			c.String(http.StatusBadRequest, "Invalid Content-Type. Expected %s", expected)
			c.Abort()
			return
		}
		c.Next()
	}
}
