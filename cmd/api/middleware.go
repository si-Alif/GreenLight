package main

import (
	"fmt"
	"net/http"

	"golang.org/x/time/rate"
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