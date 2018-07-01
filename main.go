package main

import (
	"context"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/dgrijalva/jwt-go/request"
	"github.com/gin-gonic/gin"
	"github.com/thinkerou/favicon"
	"golang.org/x/text/language"

	"github.com/recoilme/slowpoke"

	"github.com/recoilme/tgram/models"
	"github.com/recoilme/tgram/routers"
	//"github.com/thinkerou/favicon"
)

// Keep this two config private, it should not expose to open source
const NBSecretPassword = "A String Very Very Very Strong!!@##$!@#$"
const NBRandomPassword = "A String Very Very Very Niubilty!!@##$!@#4"

func main() {
	srv := &http.Server{
		Addr:    ":8081",
		Handler: InitRouter(),
	}

	go func() {
		// service connections
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server with
	// a timeout of 5 seconds.
	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt)
	<-quit
	log.Println("Shutdown Server ...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server Shutdown:", err)
	}
	// Close db
	if err := slowpoke.CloseAll(); err != nil {
		log.Fatal("Database Shutdown:", err)
	}
	log.Println("Server exiting")

}

// Strips 'TOKEN ' prefix from token string
func stripBearerPrefixFromTokenString(tok string) (string, error) {
	// Should be a bearer token
	if len(tok) > 5 && strings.ToUpper(tok[0:6]) == "TOKEN " {
		return tok[6:], nil
	}
	return tok, nil
}

// general hook
func SubDomain() gin.HandlerFunc {
	return func(c *gin.Context) {
		var lang = "ru"
		var token *jwt.Token
		var uid uint32

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
		if tokenStr, err := c.Cookie("token"); err == nil && tokenStr != "" {
			token, _ = jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
				// Don't forget to validate the alg is what you expect:
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
				}

				// hmacSampleSecret is a []byte containing your secret, e.g. []byte("my_secret_key")
				return []byte(NBSecretPassword), nil
			})
		}
		// token from header
		if token == nil {
			// parse from header
			var AuthorizationHeaderExtractor = &request.PostExtractionFilter{
				request.HeaderExtractor{"Authorization"},
				stripBearerPrefixFromTokenString,
			}
			// Extractor for OAuth2 access tokens.  Looks in 'Authorization'
			// header then 'access_token' argument for a token.
			var MyAuth2Extractor = &request.MultiExtractor{
				AuthorizationHeaderExtractor,
				request.ArgumentExtractor{"access_token"},
			}
			token, _ = request.ParseFromRequest(c.Request, MyAuth2Extractor, func(token *jwt.Token) (interface{}, error) {
				b := ([]byte(NBSecretPassword))
				return b, nil
			})
		}
		if token != nil {
			//token found
			if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
				uid = uint32(claims["id"].(float64))
			}
		}

		// set uid
		c.Set("uid", uid)
		var user models.User //users.UserModel
		_, exists := c.Get("user")
		if !exists {
			if uid > 0 {
				// get from db
			}
		}
		c.Set("user", user)

	}
}
func fmtStr(value interface{}) string {
	return fmt.Sprintf("%s", value)
}

func InitRouter() *gin.Engine {
	r := gin.Default()

	r.Use(favicon.New("./favicon.ico"))
	r.SetFuncMap(template.FuncMap{
		"fmtStr": fmtStr,
	})
	r.LoadHTMLGlob("views/*.html")

	r.Use(SubDomain())
	r.GET("/", routers.Index)
	r.GET("/register", routers.Register)
	r.POST("/register", routers.Register)
	/*
		fmt.Printf("r: %+v\n", r)


		r.GET("/register", front.Register)
		r.POST("/register", front.Register)

		r.Use(users.SetUserStatus())

		r.GET("/login", front.Login)
		r.POST("/login", front.Login)
		r.GET("/settings", front.Settings)
		r.POST("/settings", front.Settings)
		r.GET("/logout", front.Logout)
		r.GET("/editor", front.Editor)
		r.POST("/editor", front.Editor)
		r.GET("/article/:slug", front.ArticleGet)
		r.POST("/article/:slug/comments", front.Comment)
		r.GET("/article/:slug/comment/:id", front.CommentDelete)
		r.GET("/", front.Index)
		//r.Use(favicon.New("./favicon.ico"))
		r.Use(CORSMiddleware())
		v1 := r.Group("/api")

		users.UsersRegister(v1.Group("/users"))

		v1.Use(users.AuthMiddleware(false))
		articles.ArticlesAnonymousRegister(v1.Group("/articles"))
		articles.TagsAnonymousRegister(v1.Group("/tags"))

		v1.Use(users.AuthMiddleware(true))
		users.UserRegister(v1.Group("/user"))
		users.ProfileRegister(v1.Group("/profiles"))

		articles.ArticlesRegister(v1.Group("/articles"))
	*/
	return r
}

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Max-Age", "86400")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE, UPDATE")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
		c.Writer.Header().Set("Access-Control-Expose-Headers", "Content-Length")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")

		if c.Request.Method == "OPTIONS" {
			//fmt.Println("OPTIONS")
			c.AbortWithStatus(200)
		} else {
			c.Next()
		}
	}
}
