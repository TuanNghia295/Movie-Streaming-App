package main

import (
	"fmt"

	"github.com/TuanNghia295/Movie-Streaming-App/server/controllers"
	"github.com/gin-gonic/gin" // Router Library
)

func main() {
	router := gin.Default()

	router.GET("/", func(ctx *gin.Context) {
		ctx.String(200, "Server is running")
	})

	router.GET("/movies", controllers.GetMovies())
	router.GET("/movies/:imdb_id", controllers.GetMovie())
	router.POST("addmovie", controllers.AddMovie())

	if err := router.Run(":8080"); err != nil {
		fmt.Println("Failed to start server", err)
	}
}
