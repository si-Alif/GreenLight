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


// define a metadata struct around pagination
type Metadata struct {
	CurrentPage int `json:"current_page",omitzero`
	PageSize int `json:"page_size",omitzero`
	FirstPage int `json:"first_page",omitzero`
	LastPage int `json:"last_page".omitzero`
	TotalRecords int `json:"total_records",omitzero`
}

func calculateMetadata(totalRecords , page , page_size int) Metadata{
	if totalRecords == 0 {
		return  Metadata{}
	}

	return  Metadata{
		CurrentPage: page,
		PageSize: page_size,
		FirstPage: 1,
		LastPage: (totalRecords + page_size - 1) / page_size,
		TotalRecords: totalRecords,
	}

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


// Pagination related :

// how many entries to return
func (f Filters) limit() int {
	return  f.PageSize
}

// how many entries to skip
func (f Filters) offset() int {
	// even though there's a risk of int overflow but we've already put validation check of entry <= 10_000_000
	return (f.Page - 1) * f.PageSize
}


