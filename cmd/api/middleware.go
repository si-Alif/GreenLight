package main

import (
	"fmt"
	"net/http"
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