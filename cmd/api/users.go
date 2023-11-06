package main

import (
	"errors"
	"net/http"
	"time"

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
	
	err = app.models.Permissions.AddForUser(user.ID, "movies:read")
	if err!=nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	// generate a new activation token after the user is created
	token, err := app.models.Tokens.New(user.ID, 3*24*time.Hour, data.ScopeActivation)
	if err!=nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	// launch a goroutine to send email to user
	// use background helper function to execute an anonymous function
	app.background(func() {
		data := map[string]any {
			"activationToken": token.Plaintext,
			"userID": user.ID,
		}
		err = app.mailer.Send(
			user.Email,
			"user_welcome.tmpl.html",
			data,
		)
		if err!=nil {
			// use app.logger to log error instead of server error
			app.logger.PrintError(err, nil)
		}
	})

	// write json response
	err = app.writeJSON(w, http.StatusAccepted, envelope{"user": user}, nil)
	if err!=nil {
		app.serverErrorResponse(w, r, err)
		return
	}
}

// activate user after signup with activationToken
func (app *application) activateUserHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		TokenPlaintext string `json:"token"`
	}
	err := app.readJSON(w, r, &input)
	if err!=nil {
		app.badRequestResponse(w, r, err)
		return
	}

	// validate plaintext token provided by user
	v := validator.New()
	if data.ValidateTokenPlainText(v, input.TokenPlaintext); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return 
	}
	// get user associated with token
	user, err := app.models.Users.GetForToken(data.ScopeActivation, input.TokenPlaintext)
	if err!=nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			v.AddError("token", "invalid or expired token")
			app.failedValidationResponse(w, r, v.Errors)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	// update user's status
	user.Activated = true
	// update in db
	err = app.models.Users.Update(user)
	if err!=nil {
		switch {
		case errors.Is(err, data.ErrEditConflict):
			app.editConflictResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}
	// delete all tokens for user if successfully activated
	err = app.models.Tokens.DeleteAllForUser(data.ScopeActivation, user.ID)
	if err!=nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	// send updated user details
	err = app.writeJSON(w, http.StatusOK, envelope{"user": user}, nil)
	if err!=nil {
		app.serverErrorResponse(w, r, err)
	}
}
