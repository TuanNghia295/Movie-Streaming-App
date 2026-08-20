package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/TuanNghia295/Movie-Streaming-App/server/database"
	"github.com/TuanNghia295/Movie-Streaming-App/server/routes"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin" // Router Library
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

func main() {
	router := gin.Default()

	router.GET("/", func(ctx *gin.Context) {
		ctx.String(200, "Server is running")
	})

	err := godotenv.Load(".env")
	if err != nil {
		log.Println("Warning: Failed to load .env file")
	}

	allowedOrigins := os.Getenv("ALLOWED_ORIGINS")
	var origins []string
	if allowedOrigins != "" {
		origins = strings.Split(allowedOrigins, ",")
		for i := range origins {
			origins[i] = strings.TrimSpace(origins[i])
			log.Println("Allowed Origin", origins[i])
		}
	} else {
		origins = []string{"http://localhost:8080"}
		log.Println("No allowed origins specified. Defaulting to http://localhost:8080")
	}

	config := cors.DefaultConfig()
	config.AllowMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"}
	config.AllowHeaders = []string{"Origin", "Content-Length", "Content-Type", "Authorization"}
	config.AllowAllOrigins = true
	config.AllowCredentials = true
	config.MaxAge = 12 * 60 * 60 // 12 hours

	router.Use(cors.New(config))
	router.Use(gin.Logger())

	var client *mongo.Client = database.Connect()
	if err := client.Ping(context.Background(), nil); err != nil {
		fmt.Println("Failed to connect to MongoDB", err)
		log.Fatalf("Failed to reach server: %v", err)
		return
	}

	// defer: có nghĩa là khi hàm main kết thúc, nó sẽ thực hiện việc đóng kết nối với MongoDB.
	// Điều này giúp đảm bảo rằng tài nguyên được giải phóng đúng cách và không gây rò rỉ bộ nhớ hoặc kết nối mở.
	// defer trong golang được sử dụng để trì hoãn việc thực hiện một hàm cho đến khi hàm bao quanh nó hoàn thành.
	//  Trong trường hợp này, nó đảm bảo rằng kết nối MongoDB sẽ được đóng khi hàm main kết thúc, bất kể có lỗi xảy ra hay không.
	defer func() {
		err := client.Disconnect(context.Background())
		if err != nil {
			log.Fatalf("Failed to disconnect from MongoDB: %v", err)
		}
	}()

	routes.PublicRoutes(router, client)
	routes.SetuptProtectedRoutes(router, client)

	if err := router.Run(":8080"); err != nil {
		fmt.Println("Failed to start server", err)
	}
}
