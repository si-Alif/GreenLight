package main

import (
	"fmt"
	"net/http"
)

func (app *application) healthcheckHandler(w http.ResponseWriter , r *http.Request){
	js := `{"status" : "available" , "environment" : "%s" , "version" : "%s" }`
	js = fmt.Sprintf(js , app.config.env , version)

	w.Header().Set("Content-Type" , "application/json")

	w.Write([]byte(js))

}