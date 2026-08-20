package routes

import (
	"github.com/TuanNghia295/Movie-Streaming-App/server/controllers"
	"github.com/gin-gonic/gin"
)

func PublicRoutes(router *gin.Engine) {
	router.GET("/movies", controllers.GetMovies())
	router.POST("/register", controllers.RegisterUser())
	router.POST("/login", controllers.LoginUser())
	router.GET("/user", controllers.UserList())
}
