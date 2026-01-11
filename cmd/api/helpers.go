// copyable

package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/julienschmidt/httprouter"
	"greenlight.si-Alif.net/internal/validator"
)

func (app *application) readIDParam(r *http.Request) (int64 , error){
	params := httprouter.ParamsFromContext(r.Context())

	id , err := strconv.ParseInt(params.ByName("id") , 10 , 64)

	if err != nil || id < 1 {
		return 0 , errors.New("invalid id parameter")
	}

	return id , nil

}

type envelope map[string]any

func (app *application) writeJSON(w http.ResponseWriter , status int , data envelope , headers http.Header) error{
	js , err := json.MarshalIndent(data , "" , "\t") /// for structured output used MarshalIndent instead of Marshal

	if err != nil {
		http.Error(w , err.Error() , http.StatusInternalServerError)
		return err
	}

	// js = append(js, '\n') --> You use append to add a new byte to the end of the slice . As the js is byte array , we've to append byte by byte to it
	js = append(js , "\n"...) // --> If you want to append a string to a byte array , you have to append it as a slice of bytes by destructuring it

	// If we figure out that there's not going to be anymore error after a certain point then it's time to write headers
	for key , val := range headers{
		w.Header()[key] = val
	}

	w.Header().Set("Content-Type" , "application/json")
	w.WriteHeader(status)
	w.Write(js)

	return nil // if everything goes well then we return nil

}


func (app *application) readJSON(w http.ResponseWriter , r *http.Request  , dst any) error{

	maxBytes := 1_048_756 // 1MB
	r.Body = http.MaxBytesReader(w , r.Body , int64(maxBytes))


	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	err := dec.Decode(dst) // first decoding the json body to dst

	if err != nil{ // if we encountered any error , start the triage(the process of sorting out which problem needs to be resolved first)
		var syntaxError *json.SyntaxError // in case of syntax error in json body
		var unmarshalTypeError *json.UnmarshalTypeError // for type mismatch
		var invalidUnmarshalError *json.InvalidUnmarshalError // for invalid decode destination
		var maxBytesError *http.MaxBytesError

		switch  {
			case errors.As(err , &syntaxError):
				return fmt.Errorf("body contains badly-formed JSON (at character %d)", syntaxError.Offset)

			case errors.Is(err , io.ErrUnexpectedEOF):
				return errors.New("body contains badly-formed JSON")

			case errors.As(err , &unmarshalTypeError):
				if unmarshalTypeError.Field != ""{
					return fmt.Errorf("body contains incorrect JSON type for field %q", unmarshalTypeError.Field)
				}
				return fmt.Errorf("body contains incorrect JSON type (at character %d)", unmarshalTypeError.Offset)

			case errors.Is(err, io.EOF):
				return errors.New("body must not be empty")

			// In-case of facing any unknown field throw an error
			case strings.HasPrefix(err.Error() , "json: unknown field ") :
				fieldName := strings.TrimPrefix(err.Error() , "json: unknown field ")
				return fmt.Errorf("body contains unknown key %s" , fieldName)

			case errors.As(err , &maxBytesError) :
				return fmt.Errorf("body must not be larger than %d bytes" , maxBytesError.Limit)

			case errors.As(err , &invalidUnmarshalError) :
				panic(err)

			default :
				return err

		}
	}

	err = dec.Decode(&struct{}{})

	if !errors.Is(err , io.EOF){
		return errors.New("body must only contain a single JSON value")
	}

	return nil

}

/*
Check if a Key exists in the query parameter :
	- In this case , a required key would be passed with a default value and the url.Values map parsed from the request query string.

If
	the required key's value exists then it would be returned as a string
else
	the default value would be returned

	*/

func (app *application) readString(qrs url.Values , key string , defaultValue string) string {
	targetVal := qrs.Get(key)

	if targetVal == "" {
		return  defaultValue
	}

	return targetVal
}

/*
readCSV() : It'll be used to parse the genres string value (in general all the key that might contain multiple values) in query parameter into a valid slice

Suppose the query is like this :  v1/movies?title=godfather&genres=crime,drama

Here the genres will be a string value when parsed via url.Values . We need to translate this into a valid slice as it should be

Parameters for readCSV will be the target key , a default value and the request url.Values map


*/

func (app *application) readCSV(qrs url.Values , key string , defaultValue []string) []string {
	targetCSV := qrs.Get(key)

	if targetCSV == "" {
		return defaultValue
	}

	return strings.Split(targetCSV , ",")

}

// found if target integer key-value exists in the query parameter
// readInt(qrs url.Values , key string , defaultValue int , v *validator.Validator) int{}

func (app *application) returnInt(qrs url.Values , key string , defaultValue int , v *validator.Validator) int {
	targetIntStr := qrs.Get(key)

	if targetIntStr == "" {
		return defaultValue
	}

	i , err := strconv.Atoi(targetIntStr)

	if err != nil {
		v.AddError(key , "must be a integer value")
		return  defaultValue
	}

	return  i

}


// helper function to execute a function passed as an argument in a background goroutine
func (app *application) background(fn func()){
	go func(){

		defer func(){
			if err := recover();err != nil{
				app.logger.Error(fmt.Sprintf("%v" , err))
			}
		}()

		fn()

	}()

}