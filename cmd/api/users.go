package main

import (
	"errors"
	"net/http"

	"github.com/anukuljoshi/greenlight/internal/data"
	"github.com/anukuljoshi/greenlight/internal/validator"
)

func (app *application) registerUserHandler(w http.ResponseWriter, r *http.Request) {
	// create anonymous struct to hold data from request body
	var input struct {
		Name 		string `json:"name"`
		Email 		string `json:"email"`
		Password 	string `json:"password"`
	}

	// parse request body into input struct
	err := app.readJSON(w, r, &input)
	if err!=nil {
		app.badRequestResponse(w, r, err)
		return
	}
	// copy data from input to User struct
	user := &data.User{
		Name: input.Name,
		Email: input.Email,
		Activated: false,
	}
	err = user.Password.Set(input.Password)
	if err!=nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	v := validator.New()
	// validate user struct and return error message to client if any
	if data.ValidateUser(v, user); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}
	// insert user data into db if valid
	err = app.models.Users.Insert(user)
	if err!=nil {
		switch {
		case errors.Is(err, data.ErrDuplicateEmail):
			v.AddError("email", "a user with this email already exists")
			app.failedValidationResponse(w, r, v.Errors)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}
	err = app.mailer.Send(
		user.Email,
		"user_welcome.tmpl.html",
		user,
	)
	if err!=nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	// write json response
	err = app.writeJSON(w, http.StatusCreated, envelope{"user": user}, nil)
	if err!=nil {
		app.serverErrorResponse(w, r, err)
		return
	}
}
