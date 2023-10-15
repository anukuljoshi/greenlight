package main

import (
	"fmt"
	"net/http"

	"golang.org/x/time/rate"
)

func (app *application) recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// create a deferred function which will run in the event of panic as Go unwinds stack
		defer func ()  {
			if err:=recover();err!=nil {
				w.Header().Set("Connection", "close")
				app.serverErrorResponse(w, r, fmt.Errorf("%s", err))
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func (app *application) rateLimit(next http.Handler) http.Handler {
	// initialize a rate limiter which allows an average of 2 req/sec
	// with maximum of 4 req in a single burst
	limiter := rate.NewLimiter(2, 4)
	// returning function is a closure which closes over the limiter
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// call limiter.Allow() to check if request is permitted
		if !limiter.Allow() {
			app.rateLimitExceeded(w, r)
			return
		}
		next.ServeHTTP(w, r)
	})
}
