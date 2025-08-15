package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (r *Rest) GetItemsList(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"artifacts": r.svc.AllItems(),
	})
}

func (r *Rest) Forge(c *gin.Context) {
	name, err := r.svc.Forge(c.Request.Context())
	if err != nil {
		c.Error(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"job_name": name})
}

func (r *Rest) Artifacts(c *gin.Context) {
	completed := c.Query("completed") == "true"
	c.JSON(http.StatusOK, gin.H{
		"artifacts": r.svc.GetArtifacts(completed),
	})
}
