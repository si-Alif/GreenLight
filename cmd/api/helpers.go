package main

import (
	"encoding/json"
	"errors"
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

func (app *application) writeJSON(w http.ResponseWriter , status int , data any , headers http.Header) error{
	js , err := json.Marshal(data)

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