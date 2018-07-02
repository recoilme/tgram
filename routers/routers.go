package routers

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/recoilme/tgram/models"
	"golang.org/x/text/language"
)

const NBSecretPassword = "A String Very Very Very Strong!!@##$!@#$"

// general hook
func CheckAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		var lang = "ru"
		var tokenStr, username, image string

		hosts := strings.Split(c.Request.Host, ".")
		var host = hosts[0]
		if host == "localhost:8081" {
			// dev
			c.Redirect(http.StatusFound, "http://sub."+host)
			return
		}
		if host == "tgr" {
			// tgr.am

			t, _, err := language.ParseAcceptLanguage(c.Request.Header.Get("Accept-Language"))
			if err == nil && len(t) > 0 {
				if len(t[0].String()) >= 2 {
					// some lang found
					if len(t[0].String()) == 2 || len(t[0].String()) == 3 {
						// some lang 3 char
						lang = t[0].String()
					} else {
						// remove country code en-US
						langs := strings.Split(t[0].String(), "-")
						if len(langs[0]) == 2 || len(langs[0]) == 3 {
							lang = langs[0]
						}
					}
				}
			}
			// redirect on subdomain
			c.Redirect(http.StatusFound, "http://"+lang+".tgr.am")
			return
		}
		if len(host) < 2 || len(host) > 3 {
			c.Redirect(http.StatusFound, "http://"+lang+".tgr.am")
			return
		}
		//fmt.Println("lang:", lang, "host:", host)
		// store subdomain
		c.Set("lang", lang)
		c.Set("path", c.Request.URL.Path)

		// token from cookie
		if tokenС, err := c.Cookie("token"); err == nil && tokenС != "" {
			tokenStr = tokenС
		}

		if tokenStr == "" {
			// token from header
			authStr := c.Request.Header.Get("Authorization")
			// Strips 'TOKEN ' prefix from token string
			if len(authStr) > 5 && strings.ToUpper(authStr[0:6]) == "TOKEN " {
				tokenStr = authStr[6:]
			}
		}
		if tokenStr != "" {
			token, tokenErr := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
				// Don't forget to validate the alg is what you expect:
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
				}

				// hmacSampleSecret is a []byte containing your secret, e.g. []byte("my_secret_key")
				return []byte(NBSecretPassword), nil
			})
			if tokenErr == nil && token != nil {
				//token found
				if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
					username = claims["username"].(string)
					image = claims["image"].(string)
				}
			}
		}
		c.Set("username", username)
		c.Set("image", image)
		/*
			var user models.User //users.UserModel
			_, exists := c.Get("user")
			if !exists {
				if uid > 0 {
					// get from db
				}
			}
			c.Set("user", user)
		*/
	}
}

func ToStr(value interface{}) string {
	return fmt.Sprintf("%s", value)
}

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

func genToken(username, image string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"username": username,
		"image":    image,
	})

	// Sign and get the complete encoded token as a string using the secret
	return token.SignedString([]byte(NBSecretPassword))
}

func Register(c *gin.Context) {
	var err error
	switch c.Request.Method {
	case "GET":
		c.HTML(http.StatusOK, "register.html", c.Keys)
	case "POST":
		var u models.User
		err = c.ShouldBind(&u)
		if err != nil {
			renderErr(c, err)
			return
		}

		// create user
		u.Lang = c.MustGet("lang").(string)
		err = models.UserNew(&u)
		if err != nil {
			renderErr(c, err)
			return
		}

		tokenString, err := genToken(u.Username, "")
		if err != nil {
			renderErr(c, err)
			return
		}

		c.SetCookie("token", tokenString, 3600, "/", "", false, true)
		switch c.Request.Header.Get("Accept") {
		case "application/json":
			c.JSON(http.StatusCreated, u)
			return
		default:
			c.Redirect(http.StatusFound, "/")
		}

	default:
		c.HTML(http.StatusNotFound, "err.html", errors.New("not found"))
	}
}

func Settings(c *gin.Context) {

	switch c.Request.Method {
	case "GET":
		user, err := models.UserGet(c.MustGet("lang").(string), c.MustGet("username").(string))
		if err != nil {
			renderErr(c, err)
			return
		}
		c.Set("bio", user.Bio)
		c.Set("email", user.Email)
		c.Set("image", user.Image)
		c.HTML(http.StatusOK, "settings.html", c.Keys)
	case "POST":
		var u models.User
		var err error
		u.Username = c.MustGet("username").(string)
		err = c.ShouldBind(&u)
		if err != nil {
			renderErr(c, err)
			return
		}
		u.Lang = c.MustGet("lang").(string)
		user, err := models.UserCheckGet(u.Lang, u.Username, u.Password)
		if err != nil {
			renderErr(c, err)
			return
		}
		//fmt.Printf("user:%+v\n", u)
		u.Password = ""
		u.PasswordHash = user.PasswordHash
		err = models.UserSave(&u)
		if err != nil {
			renderErr(c, err)
			return
		}
		if u.Image != user.Image {
			// upd token
			tokenString, err := genToken(u.Username, u.Image)
			if err != nil {
				renderErr(c, err)
				return
			}
			if c.Request.Header.Get("Accept") != "application/json" {
				c.SetCookie("token", tokenString, 3600, "/", "", false, true)
			}

		}
		switch c.Request.Header.Get("Accept") {
		case "application/json":
			c.JSON(http.StatusOK, u)
			return
		default:
			c.Redirect(http.StatusFound, "/")
		}

	}
}

func Logout(c *gin.Context) {
	switch c.Request.Method {
	case "GET":
		c.SetCookie("token", "", 0, "/", "", false, true)
		c.Redirect(http.StatusFound, "/")
	case "POST":

	}
}

func Login(c *gin.Context) {
	var err error
	switch c.Request.Method {
	case "GET":
		c.HTML(http.StatusOK, "login.html", c.Keys)
	case "POST":
		var u models.User
		err = c.ShouldBind(&u)
		if err != nil {
			renderErr(c, err)
			return
		}

		user, err := models.UserCheckGet(c.MustGet("lang").(string), u.Username, u.Password)
		if err != nil {
			renderErr(c, err)
			return
		}
		tokenString, err := genToken(user.Username, user.Image)
		if err != nil {
			renderErr(c, err)
			return
		}
		c.SetCookie("token", tokenString, 3600, "/", "", false, true)
		switch c.Request.Header.Get("Accept") {
		case "application/json":
			c.JSON(http.StatusOK, user)
		default:
			c.Redirect(http.StatusFound, "/")
		}

	}
}

func Editor(c *gin.Context) {

	log.Println("Editor", c.Request)
	switch c.Request.Method {
	case "GET":
		c.HTML(http.StatusOK, "article_edit.html", c.Keys)
	case "POST":
		x, _ := ioutil.ReadAll(c.Request.Body)
		fmt.Printf("%s", string(x))

		return
	}
}
