package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/julienschmidt/httprouter"
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
	err := json.NewDecoder(r.Body).Decode(dst)

	if err != nil{ // if we encountered any error , start the triage(the process of sorting out which problem needs to be resolved first)
		var syntaxError *json.SyntaxError // in case of syntax error in json body
		var unmarshalTypeError *json.UnmarshalTypeError // for type mismatch
		var invalidUnmarshalError *json.InvalidUnmarshalError // for invalid decode destination

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

			case errors.As(err , &invalidUnmarshalError) :
				panic(err)

			default :
				return err

		}
	}

	return nil

}