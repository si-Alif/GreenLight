package data

import (
	"context"
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

	ctx , cancel := context.WithTimeout(context.Background() , 3 * time.Second)

	defer cancel()

	return md.DB.QueryRowContext(ctx , query , args...).Scan(&movie.ID , &movie.CreatedAt , &movie.Version)
}

func (md MovieModel) Get(id int64) (*Movie , error ){
	if id < 1 {
		return  nil , ErrRecordNotFound
	}

	query := `SELECT id , created_at , title , year, runtime , genres, version FROM movies WHERE id = $1`

	var movie Movie

	ctx , cancel := context.WithTimeout(context.Background() , 3 * time.Second)

	defer cancel() // cancel the context window before we return from the Get() functions stack to prevent memory leak

	// if the query execution doesn't complete before the given time period in context then sql.DB will terminate that DB call
	err := md.DB.QueryRowContext(ctx, query , id).Scan(
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
	stmnt := "UPDATE movies SET title = $1 , year = $2 , runtime = $3 , genres = $4 , version = version + 1 WHERE id = $5 AND version=$6 RETURNING version"

	// fields to be provided for placeholder parameters as this array gets destructured
	args := []any{
		movie.Title,
		movie.Year,
		movie.Runtime,
		pq.Array(movie.Genres),
		movie.ID,
		movie.Version, // expected movie version
	}

	ctx , cancel := context.WithTimeout(context.Background() , 3 * time.Second)

	defer cancel()

	err :=  md.DB.QueryRowContext(ctx ,stmnt , args...).Scan(&movie.Version)

	if err != nil {
		switch {
			case errors.Is(err , sql.ErrNoRows):
				return ErrEditConflicts
			default:
				return  err
		}
	}

	return  nil

}

func (md MovieModel) Delete(id int64) error{
	// as delete won't perform any kind of retrieval , it's better to use Exec() in this scenario which return sql.Result object
	if id < 1 {
		return ErrRecordNotFound
	}

	stmnt := `DELETE FROM movies WHERE id=$1`

	ctx , cancel  := context.WithTimeout(context.Background() , 3 *time.Second)

	defer cancel()

	result , err := md.DB.ExecContext( ctx, stmnt , id)

	if err != nil {
		return err
	}

	rowsAffected , err := result.RowsAffected()

	if err != nil{
		return err
	}

	if rowsAffected == 0{
		return ErrRecordNotFound
	}

	return nil

}

func (md MovieModel) GetAll(title string , genres []string , filters Filters ) ([]*Movie , error){
	// --------------------------------------
	// construct a query string to retrieve all the movies data for now
	// query := `SELECT id , created_at , title , year , runtime , genres , version FROM movies ORDER BY id`
	// --------------------------------------

	/*
		✅ query for filtering parameters

		-- For EACH row in the movies table:
		-- Step 1: Get the parameter value (it's the SAME for all rows)
		parameter_title = 'black panther'  -- or '' for no filter

		-- Step 2: For each row, check:
		-- Condition A: Does this movie's title match the parameter?
		-- Condition B: Is the parameter empty?

		-- If parameter is 'black panther':
		-- Row 1 (Black Panther): ('black panther' = 'black panther') OR ('black panther' = '')
		--                       → true OR false → true ✓
		-- Row 2 (Moana):        ('moana' = 'black panther') OR ('black panther' = '')
		--                       → false OR false → false ✗

		-- If parameter is '':
		-- Row 1 (Black Panther): ('black panther' = '') OR ('' = '')
		--                       → false OR true → true ✓
		-- Row 2 (Moana):        ('moana' = '') OR ('' = '')
		--                       → false OR true → true ✓
	*/

	query := `SELECT id , created_at , title , year , runtime , genres , version FROM movies
							WHERE
								(LOWER(title) = LOWER($1) OR $1 = '') AND
								(genres @> $2 OR $2 = '{}')
							ORDER BY id
						`


	ctx , cancel := context.WithTimeout(context.Background() , 3 * time.Second)

	defer cancel()

	// retrieve resultSet of movie from database
	rows , err := md.DB.QueryContext(ctx , query , title , pq.Array(genres))

	if err != nil {
		return  nil , err
	}

	// close the connection
	defer rows.Close()

	movies := []*Movie{}

	//Use rows.Next to iterate through the resultSet
	for rows.Next() {

		// Movie instance to hold data for individual movie
		var movie Movie

		err := rows.Scan(
			&movie.ID,
			&movie.CreatedAt,
			&movie.Title,
			&movie.Year,
			&movie.Runtime,
			pq.Array(&movie.Genres),
			&movie.Version,
		)

		if err != nil {
			return  nil , err
		}

		movies = append(movies, &movie)

	}

	// check for rows.Err , does it have any error after iterating over all the resultSet ?

	if err = rows.Err() ; err != nil {
		return  nil , err
	}

	return  movies , nil

}