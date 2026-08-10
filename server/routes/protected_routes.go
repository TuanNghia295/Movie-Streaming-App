package routes

import (
	"github.com/TuanNghia295/Movie-Streaming-App/server/controllers"
	"github.com/TuanNghia295/Movie-Streaming-App/server/middleware"
	"github.com/gin-gonic/gin"
)

func SetuptProtectedRoutes(router *gin.Engine) {
	router.Use(middleware.AuthMiddleWare())

	router.GET("/movies/:imdb_id", controllers.GetMovie())
	router.POST("/addmovie", controllers.AddMovie())
	router.PATCH("/updatereview/:imdb_id", controllers.AdminReviewUpdate())

}
