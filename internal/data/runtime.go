package data

import (
	"fmt"
	"strconv"
)

// declare a custom runtime type , which will satisfy the json.Marshaler interface
type Runtime int32

func (r Runtime) MarshalJSON() ([]byte , error){
	jsonVal := fmt.Sprintf("%d mins" , r)
	quotedJSONVal := strconv.Quote(jsonVal)

	return []byte(quotedJSONVal) , nil
}
