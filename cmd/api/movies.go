package main

import (
	"errors"
	"fmt"
	"net/http"

	"greenlight.si-Alif.net/internal/data"
	"greenlight.si-Alif.net/internal/validator"
)

func (app *application) createMovieHandler(w http.ResponseWriter , r *http.Request){
	var input struct {
		Title string `json:"title"`
		Year int32 `json:"year"`
		Runtime data.Runtime `json:"runtime"` // Need to update data.Runtime so that it satisfies the json.Unmarshaler interface and decoder can use it to decode jsonValue
		Genres []string `json:"genres"`
	}

	err := app.readJSON(w , r , &input)

	if err != nil {
		app.badRequestResponse(w , r , err)
		return
	}

	// rather than manipulating or working with the input struct directly we create a new movie struct
	movie := &data.Movie{
		Title: input.Title ,
		Year: input.Year ,
		Runtime: input.Runtime ,
		Genres: input.Genres ,
	}

	v := validator.New()

	if data.ValidateMovie(v , movie) ; !v.Valid() {
		app.failedValidationResponse(w , r , v.Errors)
		return
	}


	if !v.Valid() {
		app.failedValidationResponse(w , r , v.Errors)
		return
	}

	err = app.models.Movies.Insert(movie)

	if err != nil{
		app.serverErrorResponse(w , r , err)
		return
	}

	headers := make(http.Header)

	headers.Set("Location" , fmt.Sprintf("/v1/movies/%d" , movie.ID))

	err = app.writeJSON(w , http.StatusCreated , envelope{"movie" : movie} , headers)

	if err != nil {
		app.serverErrorResponse(w , r , err)
	}

}

func (app *application) showMovieHandler(w http.ResponseWriter , r *http.Request){
	id  , err := app.readIDParam(r)

	if err != nil  {
		app.notFoundResponse(w , r)
		return
	}

	movie , err := app.models.Movies.Get(id)

	if err != nil{
		if(errors.Is(err, data.ErrRecordNotFound)){
			app.notFoundResponse(w , r)
		}else{
			app.serverErrorResponse(w , r , err)
		}
		return
	}

	err = app.writeJSON(w , http.StatusOK , envelope{"movie" : movie} , nil)
	if err != nil {
		app.serverErrorResponse(w , r , err)
	}

}


func (app *application) updateMovieHandler(w http.ResponseWriter , r *http.Request){
	id , err := app.readIDParam(r)

	if err != nil{
		app.notFoundResponse(w , r)
		return
	}

	// fetch existing movie data from database
	movie , err := app.models.Movies.Get(id)

	if err != nil{
		switch{
		case errors.Is(err , data.ErrRecordNotFound):
			app.notFoundResponse(w , r)
		default :
			app.serverErrorResponse(w , r , err)

		}
		return
	}

	// fields we expect the user to provide when they are going to use our application
	var userInput struct{
		Title string `json:"title"`
		Year int32 `json:"year"`
		Runtime data.Runtime `json:"runtime"`
		Genres []string `json:"genres"`
	}

	err = app.readJSON(w , r , &userInput)

	if err != nil{
		app.badRequestResponse(w , r , err)
		return
	}

	// update the movie fields values retrieved from database
	movie.Title = userInput.Title
	movie.Year = userInput.Year
	movie.Genres = userInput.Genres
	movie.Runtime = userInput.Runtime

	v := validator.New()

	if data.ValidateMovie(v , movie); !v.Valid() {
		app.failedValidationResponse(w , r , v.Errors)
		return
	}

	err = app.models.Movies.Update(movie)

	if err != nil{
		app.serverErrorResponse(w , r , err)
		return
	}

	err = app.writeJSON(w , http.StatusOK , envelope{"movie" : movie} , nil)

	if err != nil{
		app.serverErrorResponse(w , r , err)
	}

}

func (app *application) deleteMovieHandler(w http.ResponseWriter , r *http.Request) {
	id , err := app.readIDParam(r)

	if err != nil {
		app.notFoundResponse(w , r )
		return
	}

	err = app.models.Movies.Delete(id)

	if err != nil{
		switch {
		case errors.Is(err  ,data.ErrRecordNotFound):
			app.notFoundResponse(w ,r)
		default :
			app.serverErrorResponse(w , r ,err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"message": "movie successfully deleted"}, nil)

	if err != nil{
		app.serverErrorResponse(w , r, err)
	}

}