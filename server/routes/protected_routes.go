package routes

import (
	"github.com/TuanNghia295/Movie-Streaming-App/server/controllers"
	"github.com/TuanNghia295/Movie-Streaming-App/server/middleware"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

func SetuptProtectedRoutes(router *gin.Engine, client *mongo.Client) {
	protected := router.Group("/")
	protected.Use(middleware.AuthMiddleWare())

	protected.GET("/movies/:imdb_id", controllers.GetMovie(client))
	protected.POST("/addmovie", controllers.AddMovie(client))
	protected.PATCH("/updatereview/:imdb_id", controllers.AdminReviewUpdate(client))
	protected.GET("/recommededMovies", controllers.GetRecommendedMovies(client))
	protected.POST("/logout", controllers.Logout(client))
}
