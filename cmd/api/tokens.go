package main

import (
	"errors"
	"net/http"
	"time"

	"github.com/anukuljoshi/greenlight/internal/data"
	"github.com/anukuljoshi/greenlight/internal/validator"
)

func (app *application) createAuthenticationHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Email string `json:"email"`
		Password string `json:"password"`
	}

	err := app.readJSON(w, r, &input)
	if err!=nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	v := validator.New()
	data.ValidateEmail(v, input.Email)
	data.ValidatePasswordPlainText(v, input.Password)

	if !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	user, err := app.models.Users.GetByEmail(input.Email)
	if err!=nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.invalidCredentialsResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	// check if password matches
	match, err := user.Password.Matches(input.Password)
	if err!=nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	// call invalidCredentialsResponse if password do not match
	if !match {
		app.invalidCredentialsResponse(w, r)
		return
	}

	// generate new token with 24 hours expiry if passwords match
	token, err := app.models.Tokens.New(user.ID, 24*time.Hour, data.ScopeAuthentication)
	if err!=nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	err = app.writeJSON(w, http.StatusCreated, envelope{"token": token}, nil)
	if err!=nil {
		app.serverErrorResponse(w, r, err)
		return
	}
}
