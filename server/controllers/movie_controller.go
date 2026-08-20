package controllers

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/TuanNghia295/Movie-Streaming-App/server/database"
	"github.com/TuanNghia295/Movie-Streaming-App/server/models"
	"github.com/TuanNghia295/Movie-Streaming-App/server/utils"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/joho/godotenv"
	"github.com/tmc/langchaingo/llms/openai"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

var movieCollection *mongo.Collection = database.OpenCollection("movies")
var rankingCollection *mongo.Collection = database.OpenCollection("rankings")
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

func AdminReviewUpdate() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		role, err := utils.GetUserRoleFromContext(ctx)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Role not found in context"})
			return
		}

		if role != "ADMIN" {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "User must be part of the ADMIN role"})
			return
		}
		movieId := ctx.Param("imdb_id")

		if movieId == "" {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Movie ID required"})
			return
		}

		var req struct {
			AdminReview string `json:"admin_review"`
		}

		var resp struct {
			RankingName string `json:"ranking_name"`
			AdminReview string `json:"admin_review"`
		}

		// ShouldBind: là 1 method của gin.Context struct, nó được sử dụng để liên kết dữ liệu từ request body (thường là JSON) vào một struct trong Go.
		// Nó tự động ánh xạ các trường trong JSON với các trường tương ứng trong struct dựa trên tên và kiểu dữ liệu của chúng.
		// Nếu việc liên kết thành công, nó sẽ điền dữ liệu từ payload JSON vào struct.
		if err := ctx.ShouldBind(&req); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
			return
		}

		sentiment, rankVal, err := GetReviewRanking(req.AdminReview)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Error getting review ranking"})
			return
		}

		filter := bson.M{"imdb_id": movieId}

		update := bson.M{
			"$set": bson.M{
				"admin_review": req.AdminReview,
				"ranking": bson.M{
					"ranking_value": rankVal,
					"ranking_name":  sentiment,
				},
			},
		}

		var c, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		result, err := movieCollection.UpdateOne(c, filter, update)

		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Error updating movie"})
		}

		if result.MatchedCount == 0 {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "Movie not found"})
			return
		}

		resp.RankingName = sentiment
		resp.AdminReview = req.AdminReview

		ctx.JSON(http.StatusOK, resp)

	}
}

func GetReviewRanking(admin_review string) (string, int, error) {
	rankings, err := GetRanking()

	if err != nil {
		return "", 0, err
	}

	sentimentDelimited := ""

	for _, ranking := range rankings {
		if ranking.RankingValue != 999 {
			sentimentDelimited = sentimentDelimited + ranking.RankingName + ","
		}
	}

	sentimentDelimited = strings.Trim(sentimentDelimited, ",")

	err = godotenv.Load(".env")
	if err != nil {
		log.Println("Warning: .env file not found")
	}

	OpenAiApiKey := os.Getenv("OPENAI_API_KEY")
	if OpenAiApiKey == "" {
		return "", 0, errors.New("could not read OPENAI_API_KEY")
	}

	llm, err := openai.New(openai.WithToken(OpenAiApiKey))

	if err != nil {
		return "", 0, err
	}

	baseGromptTemplate := os.Getenv("BASE_PROMPT_TEMPLATE")
	basePrompt := strings.Replace(baseGromptTemplate, "{rankings}", sentimentDelimited, 1)

	if baseGromptTemplate == "" {
		return "", 0, errors.New("could not read BASE_PROMPT_TEMPLATE")
	}

	response, err := llm.Call(context.Background(), basePrompt+admin_review)

	if err != nil {
		return "", 0, err
	}

	rankVal := 0
	for _, ranking := range rankings {
		if ranking.RankingName == response {
			rankVal = ranking.RankingValue
			break
		}
	}

	return response, rankVal, nil
}

func GetRanking() ([]models.Ranking, error) {
	var rankings []models.Ranking

	var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)

	defer cancel()

	cursor, err := rankingCollection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}

	defer cursor.Close(ctx)
	if err := cursor.All(ctx, &rankings); err != nil {
		return nil, err
	}

	return rankings, nil
}

func GetRecommendedMovies() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		userId, err := utils.GetUserIdFromContext(ctx)

		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "User ID not found in context"})
		}

		favourite_genres, err := GetUserFavouriteGenres(userId)

		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		err = godotenv.Load(".env")
		if err != nil {
			log.Println("Warning: .env file not found")
		}

		var recommendedMovieLimitVal int64 = 5

		recommendedMovieLimitStr := os.Getenv("RECOMMENED_MOVIE_LIMIT")

		if recommendedMovieLimitStr != "" {
			recommendedMovieLimitVal, _ = strconv.ParseInt(recommendedMovieLimitStr, 10, 64)
		}

		// Create a find options and then sort the list
		findOptions := options.Find()

		findOptions.SetSort(bson.D{{Key: "ranking.ranking_value", Value: 1}})

		findOptions.SetLimit(recommendedMovieLimitVal)

		filter := bson.M{"genre.genre_name": bson.M{"$in": favourite_genres}}

		var c, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel() // run this function after 10 second

		cursor, err := movieCollection.Find(c, filter, findOptions)

		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching recommended movies"})
			return
		}

		// Clear up
		defer cursor.Close(c)

		var recommendedMovies []models.Movie

		if err := cursor.All(c, &recommendedMovies); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		ctx.JSON(http.StatusOK, recommendedMovies)
	}
}

func GetUserFavouriteGenres(userId string) ([]string, error) {
	var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel() // run this function after 10 second

	filter := bson.M{"user_id": userId}

	projection := bson.M{
		"favourite_genres.genre_name": 1,
		"_id":                         0,
	}

	// SetProjection: is a method provided by the MongoDB Go driver that allows you to specify which fields should be included or excluded in the result of a query.
	// In this case, it is used to create a projection that includes only the "favourite_genres.genre_name" field and excludes the "_id" field from the query result.

	opts := options.FindOne().SetProjection(projection)

	var result bson.M

	err := userCollection.FindOne(ctx, filter, opts).Decode(&result)

	if err != nil {
		// If the error is mongo.ErrNoDocuments, it means that no document was found matching the filter criteria.
		// In this case, we return an empty slice of strings and a nil error to indicate that there are no favorite genres for the user.
		if err == mongo.ErrNoDocuments {
			return []string{}, nil
		}
	}

	favGenresArray, ok := result["favourite_genres"].(bson.A)

	if !ok {
		return []string{}, errors.New("unable to retrieve favourite genres for user")
	}

	var genreNames []string

	for _, item := range favGenresArray {
		if genreMap, ok := item.(bson.D); ok {
			for _, elem := range genreMap {
				if elem.Key == "genre_name" {
					if name, ok := elem.Value.(string); ok {
						genreNames = append(genreNames, name)
					}
				}
			}
		}
	}

	return genreNames, nil
}
