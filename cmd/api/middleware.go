package main

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/anukuljoshi/greenlight/internal/data"
	"github.com/anukuljoshi/greenlight/internal/validator"
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

// authentication middleware
func (app *application) authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Add the "Vary: Authorization" header to the response. This indicates to any
		// caches that the response may vary based on the value of the Authorization
		// header in the request.
		w.Header().Add("Vary", "Authorization")
		// retrieve Authorization header from request
		authorizationHeader := r.Header.Get("Authorization")
		// set anonymous user if Auth header is empty
		if authorizationHeader=="" {
			r = app.contextSetUser(r, data.AnonymousUser)
			next.ServeHTTP(w, r)
			return
		}

		// extract token from Authorization header
		headerParts := strings.Split(authorizationHeader, " ")
		// check if Auth header is in correct format
		if len(headerParts)!=2 || headerParts[0]!="Bearer" {
			app.invalidAuthenticationTokenResponse(w, r)
			return
		}
		token := headerParts[1]
		// check if token is valid
		v := validator.New()
		if data.ValidateTokenPlainText(v, token); !v.Valid() {
			app.invalidAuthenticationTokenResponse(w, r)
			return
		}
		// get user with token
		user, err := app.models.Users.GetForToken(data.ScopeAuthentication, token)
		if err!=nil {
			switch {
			case errors.Is(err, data.ErrRecordNotFound):
				app.invalidAuthenticationTokenResponse(w, r)
			default:
				app.serverErrorResponse(w, r, err)
			}
			return
		}
		// set user in request context
		r = app.contextSetUser(r, user)
		// call next handler in chain
		next.ServeHTTP(w, r)
	})
}

// middle to check if user is authenticated
func (app *application) requireAuthenticatedUser(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := app.contextGetUser(r)
		if user.IsAnonymous() {
			app.authenticationRequiredResponse(w, r)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// middle to check if user is authenticated and activated
func (app *application) requireActivatedUser(next http.HandlerFunc) http.HandlerFunc {
	// instead of returning handler store in fn
	fn := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := app.contextGetUser(r)
		if !user.Activated {
			app.inactiveAccountResponse(w, r)
			return
		}
		next.ServeHTTP(w, r)
	})
	// wrap check for activated user inside requireAuthenticatedUser
	return app.requireAuthenticatedUser(fn)
}

// middle to check if user is authenticated and activated
func (app *application) requirePermission(code string, next http.HandlerFunc) http.HandlerFunc {
	// instead of returning handler store in fn
	fn := func(w http.ResponseWriter, r *http.Request) {
		user := app.contextGetUser(r)
		permissions, err := app.models.Permissions.GetAllForUser(user.ID)
		if err!=nil {
			app.serverErrorResponse(w, r, err)
			return
		}
		// check if user has required permission
		if !permissions.Include(code) {
			app.notPermittedResponse(w, r)
			return
		}
		next.ServeHTTP(w, r)
	}
	// wrap check for activated user inside requireAuthenticatedUser
	return app.requireActivatedUser(fn)
}

// middle to enable cors header
func (app *application) enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// warns caches the response may vary based on Origin header
		w.Header().Set("Vary", "Origin")

		origin := r.Header.Get("Origin")
		// set Access-Control-Allow-Origin to Origin header if it matches one of our trustedOrigins
		if origin!="" && len(app.config.cors.trustedOrigins)>0 {
			for i := range app.config.cors.trustedOrigins {
				if origin==app.config.cors.trustedOrigins[i] {
					w.Header().Set("Access-Control-Allow-Origin", origin)
				}
			}
		}
		next.ServeHTTP(w, r)
	})
}
