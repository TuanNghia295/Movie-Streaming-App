package main

import (
	"fmt"

	"github.com/gin-gonic/gin" // Router Library
)

func main() {
	router := gin.Default()

	router.GET("/", func(ctx *gin.Context) {
		ctx.String(200, "Server is running")
	})

	if err := router.Run(":8080"); err != nil {
		fmt.Println("Failed to start server", err)
	}
}
