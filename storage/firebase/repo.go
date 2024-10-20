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
	newTvShow := TVShow{
		ID:        show.ID,
		Name:      show.Name,
		Seasons:   show.Seasons,
		Episodes:  show.Episodes,
		Runtime:   show.Runtime,
		CreatedAt: time.Now(),
	}

	log.Printf("writing data: %v", newTvShow)

	_, err := getCollection(tvShowsCollection).Doc(show.ID).Set(FirestoreContext, newTvShow)
	if err != nil {
		log.Fatalf("Failed adding TV show: %v", err)
	}

	log.Println("TV show added to tvshows_dev collection successfully!")
}

// GetTvShow retrieves a TV show by its ID
func GetTvShow(id string) (*TVShow, error) {
	doc, err := getCollection(tvShowsCollection).Doc(id).Get(FirestoreContext)
	if err != nil {
		return nil, err
	}

	var show TVShow
	if err := doc.DataTo(&show); err != nil {
		return nil, err
	}
	return &show, nil
}

// UpdateTvShow updates an existing TV show
func UpdateTvShow(show *TVShow) error {
	_, err := getCollection(tvShowsCollection).Doc(show.ID).Set(FirestoreContext, show)
	if err != nil {
		return err
	}

	log.Println("TV show updated successfully!")
	return nil
}

// DeleteTvShow removes a TV show from the Firestore collection
func DeleteTvShow(id string) error {
	_, err := getCollection(tvShowsCollection).Doc(id).Delete(FirestoreContext)
	if err != nil {
		return err
	}

	log.Println("TV show deleted successfully!")
	return nil
}

// ListTvShows retrieves all TV shows from the Firestore collection
func ListTvShows() ([]TVShow, error) {
	iter := getCollection(tvShowsCollection).Documents(FirestoreContext)

	var shows []TVShow
	for {
		doc, err := iter.Next()
		if errors.Is(err, iterator.Done) {
			break
		}
		if err != nil {
			return nil, err
		}

		var show TVShow
		if err := doc.DataTo(&show); err != nil {
			return nil, err
		}
		shows = append(shows, show)
	}

	return shows, nil
}

// CreateMovie adds a new movie to the Firestore collection
func CreateMovie(movie *Movie) {
	newMovie := Movie{
		ID:       uuid.New().String(),
		Title:    movie.Title,
		Duration: movie.Duration,
	}

	_, err := getCollection(moviesCollection).Doc(newMovie.ID).Set(FirestoreContext, newMovie)
	if err != nil {
		log.Fatalf("Failed adding movie: %v", err)
	}

	log.Println("Movie added to movies_dev collection successfully!")
}

// GetMovie retrieves a movie by its ID
func GetMovie(id string) (*Movie, error) {
	doc, err := getCollection(moviesCollection).Doc(id).Get(FirestoreContext)
	if err != nil {
		return nil, err
	}

	var movie Movie
	if err := doc.DataTo(&movie); err != nil {
		return nil, err
	}
	return &movie, nil
}

// UpdateMovie updates an existing movie
func UpdateMovie(movie *Movie) error {
	_, err := getCollection(moviesCollection).Doc(movie.ID).Set(FirestoreContext, movie)
	if err != nil {
		return err
	}

	log.Println("Movie updated successfully!")
	return nil
}

// DeleteMovie removes a movie from the Firestore collection
func DeleteMovie(id string) error {
	_, err := getCollection(moviesCollection).Doc(id).Delete(FirestoreContext)
	if err != nil {
		return err
	}

	log.Println("Movie deleted successfully!")
	return nil
}

// ListMovies retrieves all movies from the Firestore collection
func ListMovies() ([]Movie, error) {
	iter := getCollection(moviesCollection).Documents(FirestoreContext)

	var movies []Movie
	for {
		doc, err := iter.Next()
		if errors.Is(err, iterator.Done) {
			break
		}
		if err != nil {
			return nil, err
		}

		var movie Movie
		if err := doc.DataTo(&movie); err != nil {
			return nil, err
		}
		movies = append(movies, movie)
	}

	return movies, nil
}
