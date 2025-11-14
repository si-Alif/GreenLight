package data

import "time"

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