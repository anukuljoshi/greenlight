package main

import (
	"net/http"
)

func (app *application) healthcheckHandler(w http.ResponseWriter, r *http.Request) {
	var data = envelope{
		"status": "available",
		"system_info": map[string]string{
			"environment": app.config.env,
			"version": version,
		},
	}
	var err = app.writeJSON(w, http.StatusOK, data, nil)
	if err!=nil {
		app.serverErrorResponse(w, r, err)
		return
	}
}
