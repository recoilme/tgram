package front

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/recoilme/tgram/common"
	"github.com/recoilme/tgram/users"
)

func Index(c *gin.Context) {
	var user users.UserModel
	iuser, uexists := c.Get("my_user_model")
	if uexists {
		user = iuser.(users.UserModel)
		log.Println("UserModel:", user)
	}

	//c.gin.H["my_user_model"] = loggedInInterface.(userModelValidator.UserModel)
	renderTemplate(c, "index", gin.H{
		"my_user_model": user,
	})
}

func Register(c *gin.Context) {
	renderTemplate(c, "register", gin.H{})
}

func Login(c *gin.Context) {
	renderTemplate(c, "login", gin.H{})
}

func LoginPost(c *gin.Context) {

	loginValidator := users.NewLoginValidator()
	if err := loginValidator.Bind(c); err != nil {
		c.HTML(http.StatusBadRequest, "login.html", gin.H{
			"ErrorTitle":   "Login Failed",
			"ErrorMessage": err.Error()})
		return
		//c.JSON(http.StatusUnprocessableEntity, common.NewValidatorError(err))
		//return
	}
	userModel, err := users.FindOneUser(&users.UserModel{Email: loginValidator.UserModel.Email})

	if err != nil {
		//c.JSON(http.StatusForbidden, common.NewError("login", errors.New("Not Registered email or invalid password")))
		c.HTML(http.StatusBadRequest, "login.html", gin.H{
			"ErrorTitle":   "Login Failed",
			"ErrorMessage": "Not Registered email or invalid password"})
		return
	}

	if userModel.CheckPassword(loginValidator.User.Password) != nil {
		c.HTML(http.StatusBadRequest, "login.html", gin.H{
			"ErrorTitle":   "Login Failed",
			"ErrorMessage": "Not Registered email or invalid password"})
		//c.JSON(http.StatusForbidden, common.NewError("login", errors.New("Not Registered email or invalid password")))
		return
	}
	users.UpdateContextUserModel(c, userModel.ID)
	c.SetCookie("token", common.GenToken(userModel.ID), 3600, "/", "", false, true)
	//serializer := UserSerializer{c}
	//c.JSON(http.StatusOK, gin.H{"user": serializer.Response()})

	//c.Set("my_user_model", userModelValidator.UserModel)
	c.Redirect(http.StatusFound, "/")
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
