package models

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

/*
Vì struct UserLogin (và cả User) không chỉ được dùng để lưu vào MongoDB, mà còn được dùng để nhận dữ liệu JSON từ HTTP request (và trả JSON về response). Đây là hai luồng dữ liệu khác nhau, mỗi luồng cần tag riêng:

json tag — dùng cho HTTP layer:
- Khi client gửi request login (thường là JSON body), Gin sẽ dùng c.BindJSON(&user) hoặc tương tự để parse JSON thành struct.
- encoding/json (và Gin bên dưới) dựa vào tag json:"..." để biết field JSON nào map với field Go nào.
- Nếu không có json tag, Go vẫn map được nhờ tên field viết hoa đầu (Email → "Email"), nhưng thường FE gửi JSON theo convention snake_case hoặc camelCase chứ không phải PascalCase, nên cần tag để match đúng key, ví dụ first_name thay vì FirstName.

bson tag — dùng cho MongoDB layer:
- Khi lưu/đọc dữ liệu từ MongoDB qua driver, driver dùng tag bson:"..." để biết cách encode/decode struct thành BSON document.
*/
type User struct {

	// struct tags are used to specify how the fields of a struct should be encoded/decoded
	// when working with BSON
	// omitempty: nếu field là zero value (0, "", nil, false...) thì sẽ bị bỏ qua khi encode.
	ID               bson.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	UserID           string        `json:"user_id" bson:"user_id"`
	FirstName        string        `json:"first_name" bson:"first_name" validate:"required,min=2,max=100"`
	LastName         string        `json:"last_name" bson:"last_name" validate:"required,min=2,max=100"`
	Email            string        `json:"email" bson:"email" validate:"required,email"`
	Password         string        `json:"password" bson:"password" validate:"required,min=6"`
	Role             string        `json:"role" bson:"role" validate:"oneof=ADMIN USER"` // 1 trong 2
	CreatedAt        time.Time     `json:"created_at" bson:"created_at"`
	UpdatedAt        time.Time     `json:"updated_at" bson:"updated_at"`
	Token            string        `json:"token" bson:"token"`
	RefreshToken     string        `json:"refresh_token" bson:"refresh_token"`
	FavouritesGenres []Genre       `json:"favourite_genres" bson:"favourite_genres" validate:"required,dive"`
}

type UserLogin struct {
	Email    string `json:"email" bson:"email" validate:"required,email"`
	Password string `json:"password" bson:"password" validate:"required,min=6"`
}

type UserResponse struct {
	UserID          string  `json:"user_id"`
	FirstName       string  `json:"first_name"`
	LastName        string  `json:"last_name"`
	Email           string  `json:"email"`
	Role            string  `json:"role"`
	Token           string  `json:"token"`
	RefreshToken    string  `json:"refresh_token"`
	FavouriteGenres []Genre `json:"favourite_genres"`
}
