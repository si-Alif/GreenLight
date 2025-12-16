package data

import (
	"strings"

	"greenlight.si-Alif.net/internal/validator"
)

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

// check if the sort query parameter value exists in our defined safe list , if it does and needs separation do it
func (f Filters) sortColumn() string {
	for _ , safeValue := range f.SortSafeList{
		if f.Sort == safeValue{
			return  strings.TrimPrefix(f.Sort , "-")
		}
	}
	panic("unsafe sort parameter" + f.Sort) // for safety purpose to prevent SQL-injection attacks
}

// figure out the orer direction based on the provided prefix in sort query parameter (eg. "-title" , "id" , "-year")
func (f Filters) sortDirection() string {
	if strings.HasPrefix(f.Sort , "-") {
		return "DESC"
	}
	return "ASC"
}