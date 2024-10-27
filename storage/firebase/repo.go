package firebase

import (
	"cloud.google.com/go/firestore"
	"errors"
	"github.com/google/uuid"
	"google.golang.org/api/iterator"
	"log"
	"time"
)

var tvShowsCollection = "tvshows_dev"
var moviesCollection = "movies_dev"

type TVShow struct {
	ID        string    `firestore:"id"`
	Name      string    `firestore:"name"`
	Seasons   int       `firestore:"seasons"`
	Episodes  int       `firestore:"episodes"`
	Runtime   int       `firestore:"runtime"`
	CreatedAt time.Time `firestore:"created_at"`
}

type Movie struct {
	ID        string    `firestore:"id"`
	Title     string    `firestore:"title"`
	Duration  int       `firestore:"duration"`
	CreatedAt time.Time `firestore:"created_at"`
}

// getCollection returns a reference to the specified Firestore collection
func getCollection(collectionName string) *firestore.CollectionRef {
	if FirestoreClient == nil {
		log.Fatalf("Firestore client is nil")
	}
	if collectionName == "" {
		log.Fatalf("Collection name is empty")
	}
	return FirestoreClient.Collection(collectionName)
}

// CreateTvShow adds a new TV show to the Firestore collection
func CreateTvShow(show *TVShow) {
	log.Printf("Request: CreateTvShow - ID: %s, Name: %s", show.ID, show.Name)

	newTvShow := TVShow{
		ID:        show.ID,
		Name:      show.Name,
		Seasons:   show.Seasons,
		Episodes:  show.Episodes,
		Runtime:   show.Runtime,
		CreatedAt: time.Now(),
	}

	_, err := getCollection(tvShowsCollection).Doc(show.ID).Set(FirestoreContext, newTvShow)
	if err != nil {
		log.Fatalf("Error: Failed to add TV show - ID: %s, Error: %v", show.ID, err)
	}
	log.Println("TV show added successfully to tvshows_dev collection")
}

// GetTvShow retrieves a TV show by its ID
func GetTvShow(id string) (*TVShow, error) {
	log.Printf("Request: GetTvShow - ID: %s", id)

	doc, err := getCollection(tvShowsCollection).Doc(id).Get(FirestoreContext)
	if err != nil {
		log.Printf("Error: Failed to retrieve TV show - ID: %s, Error: %v", id, err)
		return nil, err
	}

	var show TVShow
	if err = doc.DataTo(&show); err != nil {
		log.Printf("Error: Failed to parse TV show data - ID: %s, Error: %v", id, err)
		return nil, err
	}

	log.Printf("TV show retrieved successfully - ID: %s", show.ID)
	return &show, nil
}

// UpdateTvShow updates an existing TV show
func UpdateTvShow(show *TVShow) error {
	log.Printf("Request: UpdateTvShow - ID: %s", show.ID)

	_, err := getCollection(tvShowsCollection).Doc(show.ID).Set(FirestoreContext, show)
	if err != nil {
		log.Printf("Error: Failed to update TV show - ID: %s, Error: %v", show.ID, err)
		return err
	}

	log.Println("TV show updated successfully")
	return nil
}

// DeleteTvShow removes a TV show from the Firestore collection
func DeleteTvShow(id string) error {
	log.Printf("Request: DeleteTvShow - ID: %s", id)

	_, err := getCollection(tvShowsCollection).Doc(id).Delete(FirestoreContext)
	if err != nil {
		log.Printf("Error: Failed to delete TV show - ID: %s, Error: %v", id, err)
		return err
	}

	log.Println("TV show deleted successfully")
	return nil
}

// ListTvShows retrieves all TV shows from the Firestore collection
func ListTvShows() ([]TVShow, error) {
	log.Println("Request: ListTvShows")

	iter := getCollection(tvShowsCollection).Documents(FirestoreContext)
	var shows []TVShow
	for {
		doc, err := iter.Next()
		if errors.Is(err, iterator.Done) {
			break
		}
		if err != nil {
			log.Printf("Error: Failed to iterate over TV shows, Error: %v", err)
			return nil, err
		}

		var show TVShow
		if err := doc.DataTo(&show); err != nil {
			log.Printf("Error: Failed to parse TV show data, Error: %v", err)
			return nil, err
		}
		shows = append(shows, show)
	}

	log.Println("TV shows retrieved successfully")
	return shows, nil
}

// CreateMovie adds a new movie to the Firestore collection
func CreateMovie(movie *Movie) {
	log.Printf("Request: CreateMovie - Title: %s", movie.Title)

	newMovie := Movie{
		ID:        uuid.New().String(),
		Title:     movie.Title,
		Duration:  movie.Duration,
		CreatedAt: time.Now(),
	}

	_, err := getCollection(moviesCollection).Doc(newMovie.ID).Set(FirestoreContext, newMovie)
	if err != nil {
		log.Fatalf("Error: Failed to add movie - Title: %s, Error: %v", movie.Title, err)
	}

	log.Println("Movie added successfully to movies_dev collection")
}

// GetMovie retrieves a movie by its ID
func GetMovie(id string) (*Movie, error) {
	log.Printf("Request: GetMovie - ID: %s", id)

	doc, err := getCollection(moviesCollection).Doc(id).Get(FirestoreContext)
	if err != nil {
		log.Printf("Error: Failed to retrieve movie - ID: %s, Error: %v", id, err)
		return nil, err
	}

	var movie Movie
	if err := doc.DataTo(&movie); err != nil {
		log.Printf("Error: Failed to parse movie data - ID: %s, Error: %v", id, err)
		return nil, err
	}
	log.Printf("Movie retrieved successfully - ID: %s", movie.ID)
	return &movie, nil
}

// UpdateMovie updates an existing movie
func UpdateMovie(movie *Movie) error {
	log.Printf("Request: UpdateMovie - ID: %s", movie.ID)

	_, err := getCollection(moviesCollection).Doc(movie.ID).Set(FirestoreContext, movie)
	if err != nil {
		log.Printf("Error: Failed to update movie - ID: %s, Error: %v", movie.ID, err)
		return err
	}

	log.Println("Movie updated successfully")
	return nil
}

// DeleteMovie removes a movie from the Firestore collection
func DeleteMovie(id string) error {
	log.Printf("Request: DeleteMovie - ID: %s", id)

	_, err := getCollection(moviesCollection).Doc(id).Delete(FirestoreContext)
	if err != nil {
		log.Printf("Error: Failed to delete movie - ID: %s, Error: %v", id, err)
		return err
	}

	log.Println("Movie deleted successfully")
	return nil
}

// ListMovies retrieves all movies from the Firestore collection
func ListMovies() ([]Movie, error) {
	log.Println("Request: ListMovies")

	iter := getCollection(moviesCollection).Documents(FirestoreContext)
	var movies []Movie
	for {
		doc, err := iter.Next()
		if errors.Is(err, iterator.Done) {
			break
		}
		if err != nil {
			log.Printf("Error: Failed to iterate over movies, Error: %v", err)
			return nil, err
		}

		var movie Movie
		if err := doc.DataTo(&movie); err != nil {
			log.Printf("Error: Failed to parse movie data, Error: %v", err)
			return nil, err
		}
		movies = append(movies, movie)
	}

	log.Println("Movies retrieved successfully")
	return movies, nil
}
