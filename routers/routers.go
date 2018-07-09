package routers

import (
	"errors"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/microcosm-cc/bluemonday"
	"github.com/recoilme/tgram/models"
	"golang.org/x/text/language"
	"gopkg.in/russross/blackfriday.v2"
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

func ToDate(t time.Time) string {

	return t.Format("01.02.2006 15:04")
	//year, month, day := t.Date()
	//return fmt.Sprintf("%02d.%02d.%02d %02d:%02d", day, month, year, t.Hour, t.Minute)
}

func Home(c *gin.Context) {

	c.HTML(http.StatusOK, "home.html", c.Keys)
}

func All(c *gin.Context) {
	articles, _, _ := models.AllArticles(c.MustGet("lang").(string), "", "")
	c.Set("articles", articles)
	c.HTML(http.StatusOK, "all.html", c.Keys)
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
		if u.Image == "" {
			u.Image = "/favicon.ico"
		}
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

	switch c.Request.Method {
	case "GET":
		aid, _ := strconv.Atoi(c.Param("aid"))
		if aid > 0 {
			// check username
			aid, _ := strconv.Atoi(c.Param("aid"))
			username := c.MustGet("username").(string)
			a, err := models.ArticleGet(c.MustGet("lang").(string), username, uint32(aid))
			if err != nil {
				renderErr(c, err)
				return
			}
			c.Set("html", a.HTML)
		}
		c.HTML(http.StatusOK, "article_edit.html", c.Keys)
	case "POST":
		var err error
		if c.Request.ContentLength > 10*1024*1024 {
			renderErr(c, errors.New("Too long content >10Mb"))
			return
		}
		body, err := ioutil.ReadAll(c.Request.Body)
		if err != nil {
			renderErr(c, err)
			return
		}
		var a models.Article
		unsafe := blackfriday.Run(body)
		html := bluemonday.UGCPolicy().SanitizeBytes(unsafe)
		a.HTML = template.HTML(html)

		a.Lang = c.MustGet("lang").(string)
		a.Author = c.MustGet("username").(string)
		a.Image = c.MustGet("image").(string)
		a.CreatedAt = time.Now()
		a.Body = string(body)
		aid, err := models.ArticleNew(&a)
		if err != nil {
			renderErr(c, err)
			return
		}
		a.ID = aid
		a.Body = ""
		c.JSON(http.StatusCreated, a) //gin.H{"article": serializer.Response()})
		//c.Redirect(http.StatusFound, fmt.Sprint("@/%s/%d", a.Author, aid))
	}
}

func Article(c *gin.Context) {
	switch c.Request.Method {
	case "GET":
		lang := c.MustGet("lang").(string)
		aid, _ := strconv.Atoi(c.Param("aid"))
		username := c.Param("username")
		a, err := models.ArticleGet(lang, username, uint32(aid))
		if err != nil {
			renderErr(c, err)
			return
		}

		c.Set("article", a)
		c.Set("body", a.HTML)
		isFolow := models.IsFollowing(lang, "fol", username, c.MustGet("username").(string))
		c.Set("isfollow", isFolow)
		followcnt := models.FollowCount(lang, "fol", username)
		c.Set("followcnt", followcnt)

		// fav
		isFav := models.IsFollowing(lang, "fav", c.Param("aid"), c.MustGet("username").(string))
		c.Set("isfav", isFav)
		favcnt := models.FollowCount(lang, "fav", c.Param("aid"))
		c.Set("favcnt", favcnt)

		c.HTML(http.StatusOK, "article.html", c.Keys)
		//c.JSON(http.StatusOK, a)
	}
}

func ArticleDelete(c *gin.Context) {
	switch c.Request.Method {
	case "GET":
		aid, _ := strconv.Atoi(c.Param("aid"))
		username := c.MustGet("username").(string)
		err := models.ArticleDelete(c.MustGet("lang").(string), username, uint32(aid))
		if err != nil {
			renderErr(c, err)
			return
		}
		c.Redirect(http.StatusFound, "/")
	}
}

func Follow(c *gin.Context) {
	switch c.Request.Method {
	case "GET":
		user := c.Param("user")
		action := c.Param("action")
		err := models.Following(c.MustGet("lang").(string), "fol", user, c.MustGet("username").(string))
		if err != nil {
			renderErr(c, err)
			return
		}
		c.Redirect(http.StatusFound, action)
	}
}

func Unfollow(c *gin.Context) {
	switch c.Request.Method {
	case "GET":
		user := c.Param("user")
		action := c.Param("action")
		err := models.Unfollowing(c.MustGet("lang").(string), "fol", user, c.MustGet("username").(string))
		if err != nil {
			renderErr(c, err)
			return
		}
		c.Redirect(http.StatusFound, action)
	}
}

func Fav(c *gin.Context) {
	switch c.Request.Method {
	case "GET":
		aid := c.Param("aid")
		action := c.Param("action")
		username := c.MustGet("username").(string)

		err := models.Following(c.MustGet("lang").(string), "fav", aid, username)
		if err != nil {
			renderErr(c, err)
			return
		}
		c.Redirect(http.StatusFound, action)
	}
}
func Unfav(c *gin.Context) {
	switch c.Request.Method {
	case "GET":
		aid := c.Param("aid")
		action := c.Param("action")
		err := models.Unfollowing(c.MustGet("lang").(string), "fav", aid, c.MustGet("username").(string))
		if err != nil {
			renderErr(c, err)
			return
		}
		c.Redirect(http.StatusFound, action)
	}
}

func GoToRegister() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.MustGet("username").(string) == "" {
			c.Redirect(http.StatusFound, "/register")
		}
	}
}

func Author(c *gin.Context) {
	authorStr := c.Param("username")
	lang := c.MustGet("lang").(string)
	author, err := models.UserGet(lang, authorStr)
	if err != nil {
		renderErr(c, err)
		return
	}
	articles, _, _ := models.ArticlesAuthor(authorStr, lang, "", "")
	c.Set("articles", articles)
	c.Set("author", author)
	isFolow := models.IsFollowing(lang, "fol", authorStr, c.MustGet("username").(string))
	c.Set("isfollow", isFolow)
	followcnt := models.FollowCount(lang, "fol", authorStr)
	c.Set("followcnt", followcnt)

	// fav
	isFav := models.IsFollowing(lang, "fav", c.Param("aid"), c.MustGet("username").(string))
	c.Set("isfav", isFav)
	favcnt := models.FollowCount(lang, "fav", c.Param("aid"))
	c.Set("favcnt", favcnt)
	c.HTML(http.StatusOK, "author.html", c.Keys)
}
