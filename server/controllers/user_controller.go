package controllers

import (
	"context"
	"net/http"
	"time"

	"github.com/TuanNghia295/Movie-Streaming-App/server/database"
	"github.com/TuanNghia295/Movie-Streaming-App/server/models"
	"github.com/TuanNghia295/Movie-Streaming-App/server/utils"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"golang.org/x/crypto/bcrypt"
)

var userCollection *mongo.Collection = database.OpenCollection("users")

func HashPassword(password string) (string, error) {
	/*
		bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost) là 1 hàm trong gói bcrypt của Go, được sử dụng để tạo ra 1 hash từ password.
		Hàm này nhận vào 2 tham số:
		- password: là chuỗi password muốn hash.
		- bcrypt.DefaultCost: là 1 hằng số định nghĩa mức độ bảo mật của hash. Mức độ này càng cao thì hash càng tốn thời gian và tài nguyên để tính toán, nhưng cũng sẽ an toàn hơn. Mức độ mặc định là 10.
	*/
	HashPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	if err != nil {
		return "", err
	}

	return string(HashPassword), nil
}

func RegisterUser() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var user models.User

		if err := ctx.ShouldBindJSON(&user); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input data"})
			return
		}

		validate := validator.New()

		if err := validate.Struct(user); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Vailidation failed", "details": err.Error()})
			return
		}

		hashedPassword, err := HashPassword(user.Password)

		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		}

		var ct, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		// Defer là 1 từ khóa trong Go.
		// Nó được sử dụng để trì hoãn việc thực hiện 1 hàm cho đến khi hàm bao quanh nó hoàn toàn
		defer cancel()

		count, err := userCollection.CountDocuments(ct, bson.M{"email": user.Email})

		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check existing user"})
			return
		}
		if count > 0 {
			ctx.JSON(http.StatusConflict, gin.H{"error": "User already exist"})
			return
		}

		user.UserID = bson.NewObjectID().Hex()
		user.CreatedAt = time.Now()
		user.UpdatedAt = time.Now()
		user.Password = hashedPassword

		result, err := userCollection.InsertOne(ctx, user)

		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
			return
		}

		ctx.JSON(http.StatusCreated, result)
	}
}

func UserList() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		c, cancel := context.WithTimeout(context.Background(), 100*time.Second)

		defer cancel()

		var users []models.User

		cursor, err := userCollection.Find(c, bson.M{})

		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"Error": "Failed to fetch users collection"})
			return
		}

		// defer cursor.Close(ctx) nghĩa là: "hoãn việc đóng cursor lại, đợi đến khi hàm hiện tại (function chứa dòng này) kết thúc thì mới thực thi"
		/*

			defer là gì:
			- Từ khóa defer trong Go lên lịch cho một lời gọi hàm được thực thi ngay trước khi hàm bao quanh nó return — dù hàm kết thúc bình thường, return sớm, hay panic.
			- Các lệnh defer được thực thi theo thứ tự LIFO (vào sau ra trước) nếu có nhiều defer.
		*/
		defer cursor.Close(c)

		if err = cursor.All(c, &users); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"Error": "Failed to decode users."})
			return
		}
		ctx.JSON(http.StatusOK, users)
	}
}

func LoginUser() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var userLogin models.UserLogin // Dữ liệu mà user gửi về

		if err := ctx.ShouldBindJSON(&userLogin); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input data"})
			return
		}

		// Request, hàm này gọi quá 10 giây thì tự cancel
		var c, cancel = context.WithTimeout(context.Background(), 100*time.Second)

		defer cancel()

		var foundUser models.User

		err := userCollection.FindOne(c, bson.M{
			"email": userLogin.Email,
		}).Decode(&foundUser)

		if err != nil {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
			return
		}

		err = bcrypt.CompareHashAndPassword([]byte(foundUser.Password), []byte(userLogin.Password))
		if err != nil {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
			return
		}

		token, refreshToken, err := utils.GenerateAllTokens(foundUser.Email, foundUser.FirstName, foundUser.LastName, foundUser.Role, foundUser.UserID)

		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate tokens"})
		}

		err = utils.UpdateAllToken(foundUser.UserID, token, refreshToken)

		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update token"})
			return
		}

		ctx.SetCookie("token", token, 3600, "/", "", true, true) // secure, httpOnly
		ctx.SetCookie("refresh_token", refreshToken, 7*24*3600, "/", "", true, true)

		ctx.JSON(http.StatusOK, models.UserResponse{
			UserID:          foundUser.UserID,
			FirstName:       foundUser.FirstName,
			LastName:        foundUser.LastName,
			Email:           foundUser.Email,
			Role:            foundUser.Role,
			Token:           token,
			RefreshToken:    refreshToken,
			FavouriteGenres: foundUser.FavouritesGenres,
		})
	}
}
