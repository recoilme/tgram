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

func renderErr(c *gin.Context, err error) {
	switch c.Request.Header.Get("Accept") {
	case "application/json":
		// Respond with JSON
		c.JSON(http.StatusUnprocessableEntity, err)
	default:
		// Respond with HTML
		c.Set("err", err)
		c.HTML(http.StatusBadRequest, "err.html", c.Keys)
	}
}

func Register(c *gin.Context) {
	var err error
	switch c.Request.Method {
	case "GET":
		c.HTML(http.StatusOK, "register.html", c.Keys)
	case "POST":
		var u models.User
		err = c.ShouldBind(&u) // .ShouldBindWith(&u, binding.FormPost)
		if err != nil {
			renderErr(c, err)
		}

		// create user
		u.Lang = c.MustGet("lang").(string)
		err = models.SaveNew(&u)
		if err != nil {
			renderErr(c, err)
		} else {
			//c.Set("my_user_model", userModelValidator.UserModel)
			//c.SetCookie("token", common.GenToken(userModelValidator.UserModel.ID), 3600, "/", "", false, true)
			c.Redirect(http.StatusFound, "/")
		}
	default:

		c.HTML(http.StatusNotFound, "err.html", errors.New("not found"))
	}

}
