package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (r *Rest) Status(c *gin.Context) {
	statuses, err := r.svc.Status(c.Request.Context())
	if err != nil {
		c.Error(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": statuses})
}
