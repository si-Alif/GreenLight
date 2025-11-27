package data

import (
	"database/sql"
	"errors"
	"time"

	"github.com/lib/pq"
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


func (md MovieModel) Insert(movie *Movie) error {
	query := `INSERT INTO movies (title , year , runtime , genres) VALUES ($1 , $2 , $3 , $4) RETURNING id , created_at , version`

	// create an args array of all the placeholders in serial
	args := []any{movie.Title , movie.Year , movie.Runtime , pq.Array(movie.Genres)}

	return md.DB.QueryRow(query , args...).Scan(&movie.ID , &movie.CreatedAt , &movie.Version)
}

func (md MovieModel) Get(id int64) (*Movie , error ){
	if id < 1 {
		return  nil , ErrRecordNotFound
	}

	query := `SELECT id , created_at , title , year, runtime , genres, version FROM movies WHERE id = $1`

	var movie Movie

	err := md.DB.QueryRow(query , id).Scan(
		&movie.ID,
		&movie.CreatedAt,
		&movie.Title,
		&movie.Year,
		&movie.Runtime,
		pq.Array(&movie.Genres),
		&movie.Version,
	)

	// There might be error for no existence of desired data or something else
	if err != nil{
		if errors.Is(err , sql.ErrNoRows){
			return  nil , ErrRecordNotFound
		}else{
			return nil , err
		}
	}

	return &movie , nil

}

func (md MovieModel) Update(movie *Movie) error{
	stmnt := "UPDATE movies SET title = $1 , year = $2 , runtime = $3 , genres = $4 , version = version + 1 WHERE id = $5 RETURNING version"

	// fields to be provided for placeholder parameters as this array gets destructured
	args := []any{
		movie.Title,
		movie.Year,
		movie.Runtime,
		pq.Array(movie.Genres),
		movie.ID,
	}

	return md.DB.QueryRow(stmnt , args...).Scan(&movie.Version)

}

func (md MovieModel) Delete(id int64) error{
	return nil
}