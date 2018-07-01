package routers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/recoilme/tgram/models"
)

func Index(c *gin.Context) {
	c.HTML(http.StatusOK, "index.html", c.Keys)
}

func Register(c *gin.Context) {
	switch c.Request.Method {
	case "GET":
		c.HTML(http.StatusOK, "register.html", c.Keys)
	case "POST":
		var u models.User
		err := c.ShouldBind(&u) // .ShouldBindWith(&u, binding.FormPost)
		if err != nil {
			switch c.Request.Header.Get("Accept") {
			case "application/json":
				// Respond with JSON
				c.JSON(http.StatusUnprocessableEntity, err)
			default:
				// Respond with HTML
				c.HTML(http.StatusBadRequest, "err.html", err)
			}
		}

		// create user
		c.Set("user", u)
	default:
		c.HTML(http.StatusNotFound, "err.html", errors.New("not found"))
	}

}
