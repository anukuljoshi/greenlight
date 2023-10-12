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
	var movie = &data.Movie{
		Title: input.Title,
		Year: input.Year,
		Runtime: input.Runtime,
		Genres: input.Genres,
	}
	var v = validator.New()
	// validations
	if data.ValidateMovie(v, movie);!v.Valid() {
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
