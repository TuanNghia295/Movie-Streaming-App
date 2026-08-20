package main

import (
	"context"
	"log"
	"time"

	"github.com/TuanNghia295/Movie-Streaming-App/server/database"
	"github.com/TuanNghia295/Movie-Streaming-App/server/models"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	movieCollection := database.OpenCollection("movies")
	userCollection := database.OpenCollection("users")
	rankingCollection := database.OpenCollection("rankings")

	seedRankings(ctx, rankingCollection)
	seedMovies(ctx, movieCollection)
	seedUsers(ctx, userCollection)

	log.Println("Seed completed successfully")
}

// rankings drive the AI sentiment scoring used by PATCH /updatereview/:imdb_id.
// RankingValue 999 ("Unrated") is the default assigned to a movie before any
// admin review exists, and GetReviewRanking excludes it from the prompt choices.
func seedRankings(ctx context.Context, rankingCollection *mongo.Collection) {
	if _, err := rankingCollection.DeleteMany(ctx, bson.M{}); err != nil {
		log.Fatalf("Failed to clear rankings collection: %v", err)
	}

	rankings := []interface{}{
		models.Ranking{RankingValue: 5, RankingName: "Excellent"},
		models.Ranking{RankingValue: 4, RankingName: "Great"},
		models.Ranking{RankingValue: 3, RankingName: "Good"},
		models.Ranking{RankingValue: 2, RankingName: "Average"},
		models.Ranking{RankingValue: 1, RankingName: "Poor"},
		models.Ranking{RankingValue: 999, RankingName: "Unrated"},
	}

	result, err := rankingCollection.InsertMany(ctx, rankings)
	if err != nil {
		log.Fatalf("Failed to seed rankings: %v", err)
	}
	log.Printf("Inserted %d rankings", len(result.InsertedIDs))
}

func seedMovies(ctx context.Context, movieCollection *mongo.Collection) {
	if _, err := movieCollection.DeleteMany(ctx, bson.M{}); err != nil {
		log.Fatalf("Failed to clear movies collection: %v", err)
	}

	action := models.Genre{GenrelID: 1, GenrelName: "Action"}
	sciFi := models.Genre{GenrelID: 2, GenrelName: "Sci-Fi"}
	drama := models.Genre{GenrelID: 3, GenrelName: "Drama"}
	thriller := models.Genre{GenrelID: 4, GenrelName: "Thriller"}
	comedy := models.Genre{GenrelID: 5, GenrelName: "Comedy"}
	horror := models.Genre{GenrelID: 6, GenrelName: "Horror"}
	romance := models.Genre{GenrelID: 7, GenrelName: "Romance"}
	animation := models.Genre{GenrelID: 8, GenrelName: "Animation"}

	movies := []interface{}{
		movie("tt1375666", "Inception", "/inception.jpg", "YoHD9XEInc0",
			[]models.Genre{action, sciFi}, 5, "Excellent"),
		movie("tt0468569", "The Dark Knight", "/dark-knight.jpg", "EXeTwQWrcwY",
			[]models.Genre{action, thriller}, 5, "Excellent"),
		movie("tt0816692", "Interstellar", "/interstellar.jpg", "zSWdZVtXT7E",
			[]models.Genre{sciFi, drama}, 5, "Excellent"),
		movie("tt6751668", "Parasite", "/parasite.jpg", "5xH0HfJHsaY",
			[]models.Genre{drama, thriller}, 4, "Great"),
		movie("tt0133093", "The Matrix", "/matrix.jpg", "vKQi3bBA1y8",
			[]models.Genre{action, sciFi}, 5, "Excellent"),
		movie("tt0110357", "The Lion King", "/lion-king.jpg", "7TavVZMewpY",
			[]models.Genre{animation, drama}, 4, "Great"),
		movie("tt0109830", "Forrest Gump", "/forrest-gump.jpg", "bLvqoHBptjg",
			[]models.Genre{drama, romance}, 5, "Excellent"),
		movie("tt0088763", "Back to the Future", "/back-to-the-future.jpg", "qvsgGtivCgs",
			[]models.Genre{comedy, sciFi}, 4, "Great"),
		movie("tt0454876", "Paranormal Activity", "/paranormal-activity.jpg", "F_UxLEqz4Ig",
			[]models.Genre{horror, thriller}, 3, "Good"),
		movie("tt0332280", "The Notebook", "/the-notebook.jpg", "PZ7DrbLzXCk",
			[]models.Genre{romance, drama}, 3, "Good"),
	}

	result, err := movieCollection.InsertMany(ctx, movies)
	if err != nil {
		log.Fatalf("Failed to seed movies: %v", err)
	}
	log.Printf("Inserted %d movies", len(result.InsertedIDs))
}

func movie(imdbID, title, posterPath, youtubeID string, genres []models.Genre, rankValue int, rankName string) models.Movie {
	return models.Movie{
		ID:         bson.NewObjectID(),
		ImdbID:     imdbID,
		Title:      title,
		PosterPath: posterPath,
		YoutubeID:  youtubeID,
		Genre:      genres,
		Ranking:    models.Ranking{RankingValue: rankValue, RankingName: rankName},
	}
}

func seedUsers(ctx context.Context, userCollection *mongo.Collection) {
	if _, err := userCollection.DeleteMany(ctx, bson.M{}); err != nil {
		log.Fatalf("Failed to clear users collection: %v", err)
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("Password123"), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("Failed to hash seed password: %v", err)
	}
	now := time.Now()

	genre := func(id int, name string) []models.Genre {
		return []models.Genre{{GenrelID: id, GenrelName: name}}
	}

	users := []interface{}{
		user("Admin", "User", "admin@moviestream.com", "ADMIN", string(hashedPassword), now, genre(1, "Action")),
		user("Test", "Viewer", "viewer@moviestream.com", "USER", string(hashedPassword), now, genre(2, "Sci-Fi")),
		user("Drama", "Fan", "dramafan@moviestream.com", "USER", string(hashedPassword), now, genre(3, "Drama")),
		user("Comedy", "Lover", "comedylover@moviestream.com", "USER", string(hashedPassword), now, genre(5, "Comedy")),
		user("Horror", "Buff", "horrorbuff@moviestream.com", "USER", string(hashedPassword), now, genre(6, "Horror")),
	}

	result, err := userCollection.InsertMany(ctx, users)
	if err != nil {
		log.Fatalf("Failed to seed users: %v", err)
	}
	log.Printf("Inserted %d users (password for all: Password123)", len(result.InsertedIDs))
}

func user(firstName, lastName, email, role, hashedPassword string, now time.Time, favouriteGenres []models.Genre) models.User {
	return models.User{
		UserID:           bson.NewObjectID().Hex(),
		FirstName:        firstName,
		LastName:         lastName,
		Email:            email,
		Password:         hashedPassword,
		Role:             role,
		CreatedAt:        now,
		UpdatedAt:        now,
		FavouritesGenres: favouriteGenres,
	}
}
