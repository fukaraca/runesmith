package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (r *Rest) Status(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{})
}
