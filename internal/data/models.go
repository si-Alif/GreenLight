package data

import (
	"database/sql"
	"errors"
)

var (
	ErrRecordNotFound = errors.New("record not found")
	ErrEditConflicts = errors.New("edit Conflict")
)

// wrapper models that wraps around copies of all defined Models
type Models struct{
	Movies MovieModel
	Users UserModel
}

// NewModels takes the DB connection pool's access though the parameter and then returns a Models structs instance(not address) to work with
func NewModels(db *sql.DB) Models{
	return Models{
		Movies : MovieModel{DB: db},
		Users: UserModel{DB: db},
	}
}

