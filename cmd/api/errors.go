package main

import (
	"fmt"
	"net/http"
)

// Takes the request reference and the error and logs it
func (app *application) logError(r *http.Request , err error ){
	var(
		method = r.Method
		path  = r.URL.RequestURI()
	)

	app.logger.Error(err.Error() , "method" , method , "uri" , path)

}

// Takes response body , request , status code , message/error and returns the json response inside "error" envelope
func (app *application) errorResponse(
		w http.ResponseWriter ,
		r *http.Request ,
		status int,
		message any, // message is any cause we might need to send different types of messages as error for the client in the json body other than string
	){
	env := envelope{
		"error" : message,
	}

	// use writeJSON helper for generating JSON response
	err := app.writeJSON(w , status , env , nil)
	if err != nil {
		app.logError(r , err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}

// Takes response body , request , error and returns the json response inside "error" envelope for server side error handling
func (app *application) serverErrorResponse(w http.ResponseWriter , r *http.Request , err error){
	app.logError(r , err)

	app.errorResponse(w , r , http.StatusInternalServerError , "The server encountered a problem and could not process your request")

}

// Takes response body , request and returns the json response inside "error" envelope for not available entities
func (app *application) notFoundResponse(w http.ResponseWriter , r *http.Request){
	app.errorResponse(w , r , http.StatusNotFound , "The requested resource could not be found")
}

// Takes response body , request and returns the json response inside "error" envelope for method not allowed
func (app *application) methodNotAllowed(w http.ResponseWriter , r *http.Request){
	msg := fmt.Sprintf("The %s method is not supported for this resource" , r.Method)
	app.errorResponse(w ,r , http.StatusMethodNotAllowed , msg)
}


