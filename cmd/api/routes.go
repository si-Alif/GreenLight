package main

import (
	"expvar"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func (app *application) routes() http.Handler {
	router := httprouter.New()

	router.NotFound = http.HandlerFunc(app.notFoundResponse)
	router.MethodNotAllowed = http.HandlerFunc(app.methodNotAllowed)
	router.HandlerFunc(http.MethodGet, "/v1/healthcheck", app.healthcheckHandler)

	/*--------------------------------------------
	router.HandlerFunc(http.MethodGet , "/v1/movies" , app.requireActivatedUserMiddleware(app.listMoviesHandler))
	router.HandlerFunc(http.MethodPost , "/v1/movies" , app.requireActivatedUserMiddleware(app.createMovieHandler))
	router.HandlerFunc(http.MethodGet , "/v1/movies/:id" , app.requireActivatedUserMiddleware(app.showMovieHandler))
	// use a PATCH request rather than using PUT
	router.HandlerFunc(http.MethodPatch , "/v1/movies/:id" , app.requireActivatedUserMiddleware(app.updateMovieHandler))
	router.HandlerFunc(http.MethodDelete, "/v1/movies/:id", app.requireActivatedUserMiddleware(app.deleteMovieHandler))

	Update the middleware to permissions based pattern
	----------------------------------------------*/

	router.HandlerFunc(http.MethodGet, "/v1/movies", app.requirePermission("movies:read", app.listMoviesHandler))
	router.HandlerFunc(http.MethodPost, "/v1/movies", app.requirePermission("movies:write", app.createMovieHandler))
	router.HandlerFunc(http.MethodGet, "/v1/movies/:id", app.requirePermission("movies:read", app.showMovieHandler))
	router.HandlerFunc(http.MethodPatch, "/v1/movies/:id", app.requirePermission("movies:write", app.updateMovieHandler))
	router.HandlerFunc(http.MethodDelete, "/v1/movies/:id", app.requirePermission("movies:write", app.deleteMovieHandler))

	// user related
	router.HandlerFunc(http.MethodPost, "/v1/users", app.registerUserHandler)
	router.HandlerFunc(http.MethodPut, "/v1/users/activated", app.activeUserHandler)

	// user token
	router.HandlerFunc(http.MethodPost, "/v1/tokens/authentication", app.createAuthenticationTokenHandler)

	// Metrics
	router.Handler(http.MethodGet, "/debug/vars", expvar.Handler())

	return app.metrics((app.recoverPanic(app.enableCORS(app.rateLimit(app.authenticate(router)))))) // Wrap the router with the panic recovery middleware

}
