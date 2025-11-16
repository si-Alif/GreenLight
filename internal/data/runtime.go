package data

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)


var ErrInvalidRuntimeFormat = errors.New("invalid runtime format")


// declare a custom runtime type , which will satisfy the json.Marshaler and json.Unmarshaler interface
type Runtime int32

func (r Runtime) MarshalJSON() ([]byte , error){
	jsonVal := fmt.Sprintf("%d mins" , r)
	quotedJSONVal := strconv.Quote(jsonVal)

	return []byte(quotedJSONVal) , nil
}

func (r *Runtime) UnmarshalJSON(jsonVal []byte) error{
	unquotedJSONVal  , err := strconv.Unquote(string(jsonVal))
	if err != nil {
		return ErrInvalidRuntimeFormat
	}

	parts := strings.Split(unquotedJSONVal , " ")

	if len(parts) != 2 && parts[1] != "mins" {
		return ErrInvalidRuntimeFormat
	}

	i , err := strconv.ParseInt(parts[0] , 10 , 32)
	if err != nil {
		return ErrInvalidRuntimeFormat
	}

	*r = Runtime(i) // type conversion of i to Runtime type and then dereference the receiver and assign it there

	return nil

}