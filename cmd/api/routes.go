package main

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func (app *application) routes() http.Handler{
	router := httprouter.New()

	router.NotFound = http.HandlerFunc(app.notFoundResponse)
	router.MethodNotAllowed = http.HandlerFunc(app.methodNotAllowed)
	router.HandlerFunc(http.MethodGet , "/v1/healthcheck" , app.healthcheckHandler)

	// movies related
	router.HandlerFunc(http.MethodGet , "/v1/movies" , app.listMoviesHandler)
	router.HandlerFunc(http.MethodPost , "/v1/movies" , app.createMovieHandler)
	router.HandlerFunc(http.MethodGet , "/v1/movies/:id" , app.showMovieHandler)
	// use a PATCH request rather than using PUT
	router.HandlerFunc(http.MethodPatch , "/v1/movies/:id" , app.updateMovieHandler)
	router.HandlerFunc(http.MethodDelete, "/v1/movies/:id", app.deleteMovieHandler)

	// user related
	router.HandlerFunc(http.MethodPost , "/v1/users" , app.registerUserHandler)
	router.HandlerFunc(http.MethodPut , "/v1/users/activated" , app.activeUserHandler)

	return app.recoverPanic(app.rateLimit(router)) // Wrap the router with the panic recovery middleware

}

