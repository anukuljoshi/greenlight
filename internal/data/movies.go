package data

import (
	"database/sql"
	"time"

	"github.com/anukuljoshi/greenlight/internal/validator"
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
	return nil
}

// retrieve a movie record with id from db
func (m MovieModel) Retrieve(id int64) (*Movie, error) {
	return nil, nil
}

// update a movie record with id from db
func (m MovieModel) Update(movie *Movie) error {
	return nil
}

// delete a movie record with id from db
func (m MovieModel) Delete(id int64) error {
	return nil
}
