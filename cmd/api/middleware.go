package main

import (
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"

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
	// define a client struct which hold the limiter and last seen time for each user
	type client struct {
		limiter *rate.Limiter
		lastSeen time.Time
	}
	// create a map to hold single rate limiter per user
	var (
		mu sync.Mutex
		clients = make(map[string]*client)
	)
	// launch a background goroutine which removes old entries from clients map once every minute
	go func() {
		time.Sleep(time.Minute)
		// lock mutex to prevent rate limit check while clean up
		mu.Lock()
		// loop through all clients
		// delete their data if not seen within last three minutes
		for ip, client := range clients {
			if time.Since(client.lastSeen) > 3 * time.Minute {
				delete(clients, ip)
			}
		}
		// unlock mutex
		mu.Unlock()
	}()

	// returning function is a closure which closes over the limiter
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// only check if rate limiter is enabled
		if app.config.limiter.enabled {
			// extract clients ip address from request
			ip, _, err := net.SplitHostPort(r.RemoteAddr)
			if err!=nil {
				app.serverErrorResponse(w, r, err)
				return
			}
			// lock mutex to prevent this code from running concurrently
			mu.Lock()
			// check if client struct exists for client ip
			// create new if does not exists
			if _, found := clients[ip]; !found {
				clients[ip] = &client{
					limiter: rate.NewLimiter(
						rate.Limit(app.config.limiter.rps),
						app.config.limiter.burst,
					),
				}
			}
			// update lastSeen time
			clients[ip].lastSeen = time.Now()

			if !clients[ip].limiter.Allow() {
				mu.Unlock()
				app.rateLimitExceededResponse(w, r)
				return
			}
			// unlock mutex before call next handler
			mu.Unlock()
		}
		next.ServeHTTP(w, r)
	})
}
