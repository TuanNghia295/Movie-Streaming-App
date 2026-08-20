package routes

import (
	"github.com/TuanNghia295/Movie-Streaming-App/server/controllers"
	"github.com/TuanNghia295/Movie-Streaming-App/server/middleware"
	"github.com/gin-gonic/gin"
)

func SetuptProtectedRoutes(router *gin.Engine) {
	protected := router.Group("/")
	protected.Use(middleware.AuthMiddleWare())

	protected.GET("/movies/:imdb_id", controllers.GetMovie())
	protected.POST("/addmovie", controllers.AddMovie())
	protected.PATCH("/updatereview/:imdb_id", controllers.AdminReviewUpdate())
	protected.GET("/recommededMovies", controllers.GetRecommendedMovies())
}
