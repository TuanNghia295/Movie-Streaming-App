package routes

import (
	"github.com/TuanNghia295/Movie-Streaming-App/server/controllers"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

func PublicRoutes(router *gin.Engine, client *mongo.Client) {
	router.GET("/movies", controllers.GetMovies(client))
	router.GET("/genres", controllers.GetGenres(client))
	router.POST("/register", controllers.RegisterUser(client))
	router.POST("/login", controllers.LoginUser(client))
	router.POST("/refresh-token", controllers.RefreshTokenHandler(client))
	router.GET("/user", controllers.UserList(client))
}
