package data

import (
	"database/sql"
	"errors"
)

var (
	ErrRecordNotFound = errors.New("record not found")
	ErrEditConflicts = errors.New("edit Conflict")
)

// wrapper models that wraps around copies of all Models
type Models struct{
	Movies MovieModel
}

// NewModels takes the DB connections pool's access though the parameter and then returns a Models structs instance(not address) to work with
func NewModels(db *sql.DB) Models{
	return Models{
		Movies : MovieModel{DB: db},
	}
}

