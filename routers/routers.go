package routers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/recoilme/tgram/models"
)

func Index(c *gin.Context) {
	c.HTML(http.StatusOK, "index.html", c.Keys)
}

func Register(c *gin.Context) {
	if c.Request.Method == "GET" {
		c.HTML(http.StatusOK, "register.html", c.Keys)
	}
	if c.Request.Method == "POST" {
		var u models.User
		if err := c.ShouldBindWith(&u, binding.Query); err == nil {
			c.JSON(http.StatusOK, gin.H{"message": "Booking dates are valid!"})
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		}
	}
}
