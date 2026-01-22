package main

import (
	"errors"
	"fmt"
	"net/http"

	"greenlight.si-Alif.net/internal/data"
	"greenlight.si-Alif.net/internal/validator"
)

func (app *application) createMovieHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Title   string       `json:"title"`
		Year    int32        `json:"year"`
		Runtime data.Runtime `json:"runtime"` // Need to update data.Runtime so that it satisfies the json.Unmarshaler interface and decoder can use it to decode jsonValue
		Genres  []string     `json:"genres"`
	}

	err := app.readJSON(w, r, &input)

	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	// rather than manipulating or working with the input struct directly we create a new movie struct
	movie := &data.Movie{
		Title:   input.Title,
		Year:    input.Year,
		Runtime: input.Runtime,
		Genres:  input.Genres,
	}

	v := validator.New()

	if data.ValidateMovie(v, movie); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	if !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = app.models.Movies.Insert(movie)

	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	headers := make(http.Header)

	headers.Set("Location", fmt.Sprintf("/v1/movies/%d", movie.ID))

	err = app.writeJSON(w, http.StatusCreated, envelope{"movie": movie}, headers)

	if err != nil {
		app.serverErrorResponse(w, r, err)
	}

}

func (app *application) showMovieHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)

	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	movie, err := app.models.Movies.Get(id)

	if err != nil {
		if errors.Is(err, data.ErrRecordNotFound) {
			app.notFoundResponse(w, r)
		} else {
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"movie": movie}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}

}

func (app *application) updateMovieHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)

	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	// fetch existing movie data from database
	movie, err := app.models.Movies.Get(id)

	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)

		}
		return
	}

	// fields we expect the user to provide when they are going to use our application
	var userInput struct {
		Title   *string       `json:"title"`
		Year    *int32        `json:"year"`
		Runtime *data.Runtime `json:"runtime"`
		Genres  []string      `json:"genres"` // value of a slice is a pointer
	}

	err = app.readJSON(w, r, &userInput)

	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}
	/* Mutate the whole movie info in the DB

	// update the movie fields values retrieved from database
	movie.Title = userInput.Title
	movie.Year = userInput.Year
	movie.Genres = userInput.Genres
	movie.Runtime = userInput.Runtime

	*/

	// Partial Update
	if userInput.Title != nil {
		movie.Title = *userInput.Title
	}

	if userInput.Runtime != nil {
		movie.Runtime = *userInput.Runtime
	}

	if userInput.Year != nil {
		movie.Year = *userInput.Year
	}

	if userInput.Genres != nil {
		movie.Genres = userInput.Genres // no need to dereference a slice cause it's already transported via a pointer and the compiler dereferences it
	}

	v := validator.New()

	if data.ValidateMovie(v, movie); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = app.models.Movies.Update(movie)

	if err != nil {
		switch {
		case errors.Is(err, data.ErrEditConflicts):
			app.editConflictResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return // whatever the error is return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"movie": movie}, nil)

	if err != nil {
		app.serverErrorResponse(w, r, err)
	}

}

func (app *application) deleteMovieHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)

	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	err = app.models.Movies.Delete(id)

	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"message": "movie successfully deleted"}, nil)

	if err != nil {
		app.serverErrorResponse(w, r, err)
	}

}

func (app *application) listMoviesHandler(w http.ResponseWriter, r *http.Request) {
	var possibleQueryParameterStruct struct {
		Title  string
		Genres []string
		data.Filters
	}

	v := validator.New()

	// generate url.Values map
	qrs := r.URL.Query()

	possibleQueryParameterStruct.Title = app.readString(qrs, "title", "")

	possibleQueryParameterStruct.Genres = app.readCSV(qrs, "genres", []string{})

	// read the required page number , fall back to 1 if nil
	possibleQueryParameterStruct.Filters.Page = app.returnInt(qrs, "page", 1, v)

	// read pagesize if provided(maybe how many entries would be rendered in a single page) . In this case default is 20
	possibleQueryParameterStruct.Filters.PageSize = app.returnInt(qrs, "page_size", 20, v)

	// if the base for sorting isn't provided , it'll fallback to id in ascending order
	possibleQueryParameterStruct.Filters.Sort = app.readString(qrs, "sort", "id")

	// define all the sort safeValue list , it'll be compared with the sort query parameter provided by the user
	possibleQueryParameterStruct.Filters.SortSafeList = []string{"id", "title", "year", "runtime", "-id", "-title", "-year", "-runtime"}

	// for all the query parameters of listMoviesHandler , a centralized validator instance is used all over
	if data.ValidateFilters(v, possibleQueryParameterStruct.Filters); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	movies, metadata, err := app.models.Movies.GetAll(
		possibleQueryParameterStruct.Title,
		possibleQueryParameterStruct.Genres,
		possibleQueryParameterStruct.Filters,
	)

	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"movies": movies, "metadata": metadata}, nil)

	if err != nil {
		app.serverErrorResponse(w, r, err)
	}

}
