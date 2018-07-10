package main

import (
	"context"
	"html/template"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
	"github.com/thinkerou/favicon"

	"github.com/recoilme/slowpoke"

	"github.com/recoilme/tgram/routers"
	//"github.com/thinkerou/favicon"
)

// Keep this two config private, it should not expose to open source

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

func InitRouter() *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New() // .Default()
	r.Use(gin.Recovery())
	//r.Use(gin.Logger())
	//gin.DefaultWriter = ioutil.Discard

	r.Use(favicon.New("./favicon.ico"))
	r.Use(static.Serve("/m", static.LocalFile("./media", false)))

	r.SetFuncMap(template.FuncMap{
		"tostr":  routers.ToStr,
		"todate": routers.ToDate,
	})
	r.LoadHTMLGlob("views/*.html")

	r.Use(routers.CheckAuth())
	r.GET("/", routers.Home)
	r.GET("/all", routers.All)

	r.GET("/register", routers.Register)
	r.POST("/register", routers.Register)

	r.GET("/login", routers.Login)
	r.POST("/login", routers.Login)

	r.GET("/@:username/:aid", routers.Article)
	r.GET("/@:username", routers.Author)

	r.Use(routers.GoToRegister())

	r.GET("/settings", routers.Settings)
	r.POST("/settings", routers.Settings)

	r.GET("/logout", routers.Logout)

	r.GET("/delete/a/:aid", routers.ArticleDelete)

	r.GET("/editor/:aid", routers.Editor)
	r.POST("/editor/:aid", routers.Editor)

	r.GET("follow/:user/*action", routers.Follow)
	r.GET("unfollow/:user/*action", routers.Unfollow)

	r.GET("fav/:aid/*action", routers.Fav)
	r.GET("unfav/:aid/*action", routers.Unfav)

	r.POST("/comments/@:username/:aid", routers.CommentNew)
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
