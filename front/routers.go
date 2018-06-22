package front

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/recoilme/tgram/users"
)

func Index(c *gin.Context) {
	renderTemplate(c, "index", gin.H{
		"title": "Home",
	})
}

func Register(c *gin.Context) {
	renderTemplate(c, "register", gin.H{})
}

func RegisterPost(c *gin.Context) {
	userModelValidator := users.NewUserModelValidator()

	if err := userModelValidator.Bind(c); err != nil {
		c.HTML(http.StatusBadRequest, "register.html", gin.H{
			"ErrorTitle":   "Registration Failed",
			"ErrorMessage": err.Error()})
		return
	}
	if err := users.SaveOne(&userModelValidator.UserModel); err != nil {
		c.HTML(http.StatusBadRequest, "register.html", gin.H{
			"ErrorTitle":   "database",
			"ErrorMessage": err.Error()})
		return
	}
	c.Set("my_user_model", userModelValidator.UserModel)
	c.Redirect(http.StatusFound, "/")
}

func renderTemplate(c *gin.Context, tmpl string, p interface{}) {
	c.HTML(http.StatusOK, tmpl+".html", p)
}
