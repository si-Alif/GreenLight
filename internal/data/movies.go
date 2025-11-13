package data

import "time"

// for holding all the relevant data about a movie
type Movie struct {
	ID int64
	CreatedAt time.Time
	Title string
	Year int32
	Runtime int32 // to store the runtime in minutes
	Genres []string
	Version int32 //starts with 1 and then gets incremented each time the movie info is updated
}