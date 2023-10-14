package data

import (
	"database/sql"
	"errors"
	"time"

	"github.com/anukuljoshi/greenlight/internal/validator"
	"github.com/lib/pq"
)

type Movie struct {
	ID 			int64		`json:"id"`
	CreatedAt 	time.Time	`json:"-"`
	Title 		string		`json:"title"`
	Year 		int32		`json:"year,omitempty"`
	Runtime 	Runtime		`json:"runtime,omitempty"`
	Genres 		[]string	`json:"genres,omitempty"`
	Version 	int32		`json:"version"`
}

func ValidateMovie(v *validator.Validator, movie *Movie) {
	// title
	v.Check(movie.Title!="", "title", "required")
	v.Check(len(movie.Title)<=500, "title", "must not be more than 500 characters")

	// year
	v.Check(movie.Year!=0, "year", "required")
	v.Check(movie.Year>=1888, "year", "must be greater than 1888")
	v.Check(movie.Year<=int32(time.Now().Year()), "year", "must not be in the future")

	// runtime
	v.Check(movie.Runtime!=0, "runtime", "required")
	v.Check(movie.Runtime>0, "runtime", "must be a positive integer")

	// genres
	v.Check(movie.Genres!=nil, "genres", "required")
	v.Check(len(movie.Genres)>=1, "genres", "must contain at least one genre")
	v.Check(len(movie.Genres)<=5, "genres", "must not contain more than 5 genres")
	v.Check(validator.Unique(movie.Genres), "genres", "must contain unique values")
}

type MovieModel struct {
	DB *sql.DB
}

// create a movie instance in db
func (m MovieModel) Create(movie *Movie) error {
	query := `
		INSERT INTO movies (title, year, runtime, genres)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at, version
	`
	args := []any{movie.Title, movie.Year, movie.Runtime, pq.Array(movie.Genres)}
	return m.DB.QueryRow(query, args...).Scan(&movie.ID, &movie.CreatedAt, &movie.Version)
}

// retrieve a movie record with id from db
func (m MovieModel) Get(id int64) (*Movie, error) {
	if id<1 {
		return nil, ErrRecordNotFound
	}
	query := `
		SELECT id, title, year, runtime, genres, created_at, version
		FROM movies
		WHERE id = $1;
	`
	var movie Movie
	err := m.DB.QueryRow(query, id).Scan(
		&movie.ID,
		&movie.Title,
		&movie.Year,
		&movie.Runtime,
		pq.Array(&movie.Genres),
		&movie.CreatedAt,
		&movie.Version,
	)
	if err!=nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}
	return &movie, nil
}

// update a movie record with id from db
func (m MovieModel) Update(movie *Movie) error {
	query := `
		UPDATE movies
		SET title = $1, year = $2, runtime = $3, genres = $4, version = version + 1
		WHERE id = $5
		RETURNING version;
	`
	args := []any{movie.Title, movie.Year, movie.Runtime, pq.Array(movie.Genres), movie.ID}
	return m.DB.QueryRow(query, args...).Scan(&movie.Version)
}

// delete a movie record with id from db
func (m MovieModel) Delete(id int64) error {
	return nil
}
