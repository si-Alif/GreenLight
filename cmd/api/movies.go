package main

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/julienschmidt/httprouter"
)

func (app *application) createMovieHandler(w http.ResponseWriter , r *http.Request){
	fmt.Fprintln(w , "create a movie")
}

func (app *application) showMovieHandler(w http.ResponseWriter , r *http.Request){
	params := httprouter.ParamsFromContext(r.Context())

	id , err :=	strconv.ParseInt(params.ByName("id") , 10 , 64)

	if err != nil {
		http.NotFound(w , r)
		return
	}

	fmt.Fprintf(w , "show the movie details of movie %d\n" , id)

}