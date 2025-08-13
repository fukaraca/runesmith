package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func NoRoute404(c *gin.Context) {
	c.JSON(http.StatusBadRequest, gin.H{"fuck": "off"})
}
