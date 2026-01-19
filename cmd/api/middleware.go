package main

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/tomasen/realip"
	"golang.org/x/time/rate"
	"greenlight.si-Alif.net/internal/data"
	"greenlight.si-Alif.net/internal/validator"
)

func (app *application) recoverPanic(next http.Handler) http.Handler{
	return http.HandlerFunc(func(w http.ResponseWriter , r *http.Request){
		// create a deferred function which will run constantly while unwinding error stack
		defer func ()  {
			if err := recover(); err != nil {
				w.Header().Set("Connection" , "close") // close the connection once error occurs
				app.serverErrorResponse(w , r , fmt.Errorf("%s" , err))
			}
		}()

		next.ServeHTTP(w , r)
	})
}

/*--------------------------
	1️⃣ Global rate limiting

func (app *application) rateLimit(next http.Handler) http.Handler{

limiter := rate.NewLimiter(2 , 4) // bucket size of 4 with refilling rate of 2tokens/s

// return a closure functions , it closes over the limiter variable defined above
return http.HandlerFunc(func(w http.ResponseWriter , r *http.Request){
	if !limiter.Allow() {
		app.rateLimitExceededResponse(w , r)
		return
	}

	next.ServeHTTP(w , r)
})

}

*/


// IP-based rate-limiting
func (app *application) rateLimit(next http.Handler) http.Handler{

	if !app.config.limiter.enabled {
		return next // return if no limiter is needed
	}



	// define a client struct to hold the rate limiter and last seen time
	type client struct {
		limiter rate.Limiter
		lastSeen time.Time
	}

	var (
		mu sync.Mutex // centralized mutex for rate-limiting purpose
		clients = make(map[string]*client)
	)

	// background goroutine for cleaning up old entries (per minute check and every 3/mins of inconsistency , cleanup)
	go func(){
		// infinite loop
		for {
			time.Sleep(time.Minute)

			// lock the mutex while the clean
			mu.Lock()

			for ip , client := range clients{
				if time.Since(client.lastSeen) > 3 * time.Minute {
					delete(clients , ip)
				}
			}

			mu.Unlock()
		}
	}() //

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// retrieve the IP address from the request
		ip := realip.FromRequest(r)

		// lock the mutex to prevent race conditions by concurrent code
		mu.Lock()

		// check if the client exists in the map and if not add their IP with a rate limiter instance
		if _ , found := clients[ip]; !found {
			clients[ip] = &client{
				limiter:*rate.NewLimiter(rate.Limit(app.config.limiter.rps) , app.config.limiter.burst)}
		}

		// once the limiter been assigned , attach current time
		clients[ip].lastSeen = time.Now()

		if !clients[ip].limiter.Allow(){
			mu.Unlock()
			app.rateLimitExceededResponse(w , r)
			return
		}

		mu.Unlock()

		next.ServeHTTP(w , r)

	})

}


// authentication middleware
func (app *application) authenticate(next http.Handler) http.Handler{
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// Add "Vary: Authorization" to indicate caches that based on this value response might vary
		w.Header().Add("Vary" , "Authorization")

		authorizationHeader := r.Header.Get("Authorization")

		if authorizationHeader == ""{
			r = app.SetUserInRequestContext(r , data.AnonymousUser)
			next.ServeHTTP(w , r) // call the next handler in the chain
			return
		}

		headerParts := strings.Split(authorizationHeader, " ")
		if len(headerParts) != 2 || headerParts[0] != "Bearer"{
			app.invalidAuthenticationTokenResponse(w , r )
			return
		}

		token := headerParts[1]

		v := validator.New()

		if data.ValidPlainTextToken(v , token); !v.Valid(){
			app.invalidAuthenticationTokenResponse(w , r)
			return
		}

		user , err := app.models.Users.GetUserViaToken(data.ScopeAuthentication , token)
		if err != nil {
			switch {
				case errors.Is(err , data.ErrRecordNotFound):
					app.invalidAuthenticationTokenResponse(w , r)
				default :
					app.serverErrorResponse(w , r , err)

			}
			return
		}


		r = app.SetUserInRequestContext(r , user)

		// pass the authority to next handler in the chain if the user has been set in the subsequent request successfully
		next.ServeHTTP(w , r)
	})
}

// check if a user is authenticated
func (app *application) requireAuthenticatedUserMiddleware(next http.HandlerFunc) http.HandlerFunc{
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := app.GetUserFromSubsequentRequestContext(r)

		if user.IsAnonymous(){
			app.AuthenticationRequiredResponse(w , r)
			return
		}

		next.ServeHTTP(w ,r)
	})
}

// check if the user is both authenticated and authorized to perform
func (app *application) requireActivatedUserMiddleware(next http.HandlerFunc) http.HandlerFunc{
	// create the authorization checker middleware
	fn := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := app.GetUserFromSubsequentRequestContext(r)

		if !user.Activated{
			app.ActivationRequiredResponse(w ,r)
			return
		}

		next.ServeHTTP(w ,r)
	})


	return app.requireAuthenticatedUserMiddleware(fn)

}

func (app *application) requirePermission(code string , next http.HandlerFunc) http.HandlerFunc{
	fn := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := app.GetUserFromSubsequentRequestContext(r)

		permissions , err := app.models.Permissions.GetAllPermissionsForUser(user.ID)
		if err != nil {
			app.serverErrorResponse(w ,r ,err)
			return
		}

		if !permissions.Include(code){
			app.notPermittedResponse(w ,r)
			return
		}

		next.ServeHTTP(w ,r)
	})

	return  app.requireActivatedUserMiddleware(fn)
}

func (app *application) enableCORS(next http.Handler) http.Handler{
	return  http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// response differs based on Origin , so it should be included in the
		w.Header().Set("Vary" , "Origin")

		// A preflight request would have 3 components : HTTP method would be Options , Access-Control-Request-Method , Origin Header

		w.Header().Add("Vary" , "Access-Control-Request-Method")

		// 1 . We retrieved the Origin here
		origin := r.Header.Get("Origin")

		// check if origin is among one of the trusted origin
		if origin != ""{
			for i :=range app.config.cors.trustedOrigins{
				if origin == app.config.cors.trustedOrigins[i]{
					w.Header().Set("Access-Control-Allow-Origin" , origin)

					// 2. Checking if the request has Options method and Access-Control-Request-Method header . If it satisfies , it's a preflight request
					if r.Method == http.MethodOptions && r.Header.Get("Access-Control-Request-Method") != ""{
						w.Header().Set("Access-Control-Request-Methods" , "OPTIONS, PUT, PATCH, DELETE")
						w.Header().Set("Access-Control-Allow-Headers" , "Authorization, Content-type")
						w.WriteHeader(http.StatusOK) // if the preflight request has been processed successfully , return preflight cors request response and return
						return
					}
					break
				}
			}
		}

		next.ServeHTTP(w ,r)
	})
}