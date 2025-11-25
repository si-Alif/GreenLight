package data

import (
	"database/sql"
	"time"

	"greenlight.si-Alif.net/internal/validator"
)

// for holding all the relevant data about a movie

// next step : modified movie struct to show snake case when the appear as a response
type Movie struct {
	ID int64 `json:"id"`
	CreatedAt time.Time `json:"-"` // used - directive so that this field is omitted from the response body
	Title string `json:"title"`
	Year int32 `json:"year,omitzero"`
	Runtime Runtime  `json:"runtime,omitzero"`      // to store the runtime in minutes
	Genres []string `json:"genres,omitzero"`
	Version int32  `json:"version"`   //starts with 1 and then gets incremented each time the movie info is updated
}


func ValidateMovie(v *validator.Validator , movie *Movie){
	v.Check(movie.Title != "" , "title" , "must be provided")
	v.Check(len(movie.Title) <= 500 , "title" , "must not be more than 500 bytes long")

	v.Check(movie.Year != 0 , "year" , "must be provided")
	v.Check(movie.Year >= 1888 , "year" , "must be greater than 1888")
	v.Check(movie.Year <= int32(time.Now().Year()) , "year", "must not be in the future")

	v.Check(movie.Runtime != 0, "runtime", "must be provided")
	v.Check(movie.Runtime > 0, "runtime", "must be a positive integer")

	v.Check(movie.Genres != nil, "genres", "must be provided")
	v.Check(len(movie.Genres) >= 1, "genres", "must contain at least 1 genre")
	v.Check(len(movie.Genres) <= 5, "genres", "must not contain more than 5 genres")
	v.Check(validator.Unique(movie.Genres) , "genres", "must not contain duplicate values")

}


type MovieModel struct{
	DB *sql.DB
}


func (md MovieModel) Insert (movie *Movie) error {
	return nil
}

func (md MovieModel) Get(id int64) (*Movie , error ){
	return nil  ,nil
}

func (md MovieModel) Update(movie *Movie) error{
	return  nil
}

func (md MovieModel) Delete(id int64) error {
	return nil
}