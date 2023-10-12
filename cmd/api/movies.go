package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/anukuljoshi/greenlight/internal/data"
	"github.com/anukuljoshi/greenlight/internal/validator"
)

func (app *application) createMovieHandler(w http.ResponseWriter, r *http.Request) {
	// create anonymous struct to hold request body info
	var input struct {
		Title string `json:"title"`
		Year int32 `json:"year"`
		Runtime data.Runtime `json:"runtime"`
		Genres []string `json:"genres"`
	}
	// initialize json.Decoder() which reads from request.Body
	var err = app.readJSON(w, r, &input)
	if err!=nil {
		app.badRequestResponse(w, r, err)
		return
	}
	var v = validator.New()
	// validations
	// title
	v.Check(input.Title!="", "title", "required")
	v.Check(len(input.Title)<=500, "title", "must not be more than 500 characters")

	// year
	v.Check(input.Year!=0, "year", "required")
	v.Check(input.Year>=1888, "year", "must be greater than 1888")
	v.Check(input.Year<=int32(time.Now().Year()), "year", "must not be in the future")

	// runtime
	v.Check(input.Runtime!=0, "runtime", "required")
	v.Check(input.Runtime>0, "runtime", "must be a positive integer")

	// genres
	v.Check(input.Genres!=nil, "genres", "required")
	v.Check(len(input.Genres)>=1, "genres", "must contain at least one genre")
	v.Check(len(input.Genres)<=5, "genres", "must not contain more than 5 genres")
	v.Check(validator.Unique(input.Genres), "genres", "must contain unique values")

	if !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}
	// dump contents of input struct into http response
	fmt.Fprintf(w, "%+v\n", input)
}

func (app *application) showMovieHandler(w http.ResponseWriter, r *http.Request) {
	var id, err = app.readIDParam(r)
	if err!=nil {
		app.notFoundResponse(w, r)
		return
	}
	var movie = data.Movie{
		ID: id,
		CreatedAt: time.Now(),
		Title: "Casablanca",
		Runtime: 102,
		Genres: []string{"drama", "romance", "war"},
		Version: 1,
	}
	err = app.writeJSON(w, http.StatusOK, envelope{"movie": movie}, nil)
	if err!=nil {
		app.serverErrorResponse(w, r, err)
		return
	}
}
