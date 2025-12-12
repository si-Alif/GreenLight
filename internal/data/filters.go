package data

import "greenlight.si-Alif.net/internal/validator"

type Filters struct {
	Page int
	PageSize int
	Sort string
	SortSafeList []string
}

// add validation checks for filtering related data fields

func ValidateFilters (v *validator.Validator , f Filters){
	v.Check(f.Page > 0 , "page" , "must be greater than zero")
	v.Check(f.Page <= 10_000_000 , "page" , "must be a maximum of 10 million")

	v.Check(f.PageSize > 0 , "page_size" , "must be greater than zero")
	v.Check(f.PageSize <= 100 , "page_size" , "must be a maximum of 100")

	// check for sort parameter's value is in permitted range
	v.Check(validator.PermittedValue(f.Sort , f.SortSafeList...) , "sort" , "invalid sort value")

}