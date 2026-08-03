package controllers

import (
	"context"
	"net/http"
	"time"

	"github.com/TuanNghia295/Movie-Streaming-App/server/database"
	"github.com/TuanNghia295/Movie-Streaming-App/server/models"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

var movieCollection *mongo.Collection = database.OpenCollection("movies")
var validate = validator.New()

func GetMovies() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)

		defer cancel()

		var movies []models.Movie

		// cursor là 1 con trỏ đến 1 tập hợp các kết quả trả về từ MongoDB
		cursor, err := movieCollection.Find(ctx, bson.M{})

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"Error": "Failed to fetch movies."})
			return
		}

		// Defer là 1 từ khóa trong Go.
		// Nó được sử dụng để trì hoãn việc thực hiện 1 hàm cho đến khi hàm bao quanh nó hoàn toàn
		defer cursor.Close(ctx)

		if err = cursor.All(ctx, &movies); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"Error": "Failed to decode movies."})
			return
		}
		c.JSON(http.StatusOK, movies)
	}
}

func GetMovie() gin.HandlerFunc {
	return func(c *gin.Context) {
		//  Nếu thao tác dùng ctx này chạy quá 100s mà chưa xong, nó sẽ tự động bị hủy
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)

		// Đảm bảo cancel() luôn được gọi khi handler kết thúc (dù thành công, lỗi, panic...), để giải phóng tài nguyên gắn với context
		defer cancel()

		movieID := c.Param("imdb_id")

		if movieID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Movie ID is required"})
			return
		}

		var movie models.Movie

		err := movieCollection.FindOne(ctx, bson.M{"imdb_id": movieID}).Decode(&movie)

		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Movie not found"})
			return
		}

		c.JSON(http.StatusOK, movie)

	}
}

func AddMovie() gin.HandlerFunc {
	return func(c *gin.Context) {
		// If this interaction run more than 10 second, it will be canceled
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)

		// Make sure cancel() always be called when handler end eventhough success or error. To reduce resource and context. It prevent leak data
		defer cancel()

		var movie models.Movie // movie have type Movie from model

		// binds:là liên kết, ràng buộc
		// corresponding: tương ứng
		// ShouldBindJSON is a method of the gin.Context struct that binds the JSON payload from the request body to a Go struct.
		// It automatically maps the JSON fields to the corresponding struct fields base on their names and types.
		// If the binding is successful, it populates the struct with the data from the JSON payload.
		err := c.ShouldBindJSON(&movie)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
			return
		}

		if err := validate.Struct(movie); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Validation failed", "details": err.Error()})
			return
		}

		result, err := movieCollection.InsertOne(ctx, movie)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add movie"})
			return
		}

		c.JSON(http.StatusOK, result)

	}
}
