package main

import (
	"errors"
	"fmt"
	"net/http"

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
	// call Create method for Movie model with a pointer to a movie struct
	err = app.models.Movies.Create(movie)
	if err!=nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	// set Location with url of newly created record in headers
	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/v1/movies/%d", movie.ID))

	// return response with StatusCreated
	err = app.writeJSON(w, http.StatusCreated, envelope{"movie": movie}, headers)
	if err!=nil {
		app.serverErrorResponse(w, r, err)
		return
	}
}

func (app *application) showMovieHandler(w http.ResponseWriter, r *http.Request) {
	var id, err = app.readIDParam(r)
	if err!=nil {
		app.notFoundResponse(w, r)
		return
	}
	movie, err := app.models.Movies.Get(id)
	if err!=nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}
	err = app.writeJSON(w, http.StatusOK, envelope{"movie": movie}, nil)
	if err!=nil {
		app.serverErrorResponse(w, r, err)
		return
	}
}

func (app *application) updateMovieHandler(w http.ResponseWriter, r *http.Request) {
	var id, err = app.readIDParam(r)
	if err!=nil {
		app.notFoundResponse(w, r)
		return
	}
	// get existing movie record from db
	movie, err := app.models.Movies.Get(id)
	if err!=nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	// create anonymous struct to hold request body info
	var input struct {
		Title *string `json:"title"`
		Year *int32 `json:"year"`
		Runtime *data.Runtime `json:"runtime"`
		Genres []string `json:"genres"`
	}
	// initialize json.Decoder() which reads from request.Body
	err = app.readJSON(w, r, &input)
	if err!=nil {
		app.badRequestResponse(w, r, err)
		return
	}
	// copy values from input data to fetched movie record
	if input.Title!=nil {
		movie.Title = *input.Title
	}
	if input.Year!=nil {
		movie.Year = *input.Year
	}
	if input.Runtime!=nil {
		movie.Runtime = *input.Runtime
	}
	if input.Genres!=nil {
		movie.Genres= input.Genres
	}

	var v = validator.New()
	// validations
	if data.ValidateMovie(v, movie);!v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}
	// call Update method for Movie model with a pointer to updated movie struct
	err = app.models.Movies.Update(movie)
	if err!=nil {
		switch {
		case errors.Is(err, data.ErrEditConflict):
			app.editConflictResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	// return response with StatusCreated
	err = app.writeJSON(w, http.StatusCreated, envelope{"movie": movie}, nil)
	if err!=nil {
		app.serverErrorResponse(w, r, err)
		return
	}
}

// handler to delete movie
func (app *application) deleteMovieHandler(w http.ResponseWriter, r *http.Request) {
	var id, err = app.readIDParam(r)
	if err!=nil {
		app.notFoundResponse(w, r)
		return
	}
	err = app.models.Movies.Delete(id)
	if err!=nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)			
		}
		return
	}
	err = app.writeJSON(w, http.StatusOK, envelope{"message": "movie successfully deleted"}, nil)
	if err!=nil {
		app.serverErrorResponse(w, r, err)
		return
	}
}

func (app *application) listMoviesHandler(w http.ResponseWriter, r *http.Request) {
	// define input struct to hold query param values
	var input struct {
		Title string
		Genres []string
		data.Filters
	}

	v := validator.New()
	qs := r.URL.Query()

	// use helper methods to read query param values and write into input struct
	input.Title = app.readString(qs, "title", "")
	input.Genres = app.readCSV(qs, "genres", []string{})
	input.Page = app.readInt(qs, "page", 1, v)
	input.PageSize = app.readInt(qs, "page_size", 10, v)
	input.Sort = app.readString(qs, "sort", "id")

	// check if validator is valid
	if !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}
	fmt.Fprintf(w, "%+v\n", input)
}
