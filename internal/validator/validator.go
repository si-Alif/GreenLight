package validator

import (
	"regexp"
	"slices"
)

var (
	EmailRX = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")
)

type Validator struct {
	Errors map[string]string
}


// validator instance would be limited to a function only a functions stack frames lifetime , Not scoped globally . Once that stack frame pops out of the stack , all the validations for that function is vanished . 

// New() returns a Empty Validator hashmap
func New() *Validator {
	return &Validator{
		Errors: make(map[string]string),
	}
}

// return true if there are no errors
func (v *Validator) Valid() bool {
	return len(v.Errors) == 0
}

// add an error message with Key and Value to the map[string]string
func (v *Validator) AddError(key, message string) {
	if _, exists := v.Errors[key]; !exists {
		v.Errors[key] = message
	}
}

// checks for a condition , if false then add error
func (v *Validator) Check(ok bool, key, message string) {
	if !ok {
		v.AddError(key, message)
	}
}

// return true if a value is in a certain domain
func PermittedValue[T comparable](value T, permittedValues ...T) bool {
	return slices.Contains(permittedValues, value)
}

// if there's any match between target string and the regex pattern
func Matches(val string, rx *regexp.Regexp) bool {
	return rx.MatchString(val)
}

// return True if all the values in a slice is unique
func Unique[T comparable](values []T) bool {
	uniqueVals := make(map[T]bool)

	for _, val := range values {
		uniqueVals[val] = true
	}

	return len(uniqueVals) == len(values)
}
