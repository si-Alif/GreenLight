package main

import (
	"net/http"
)

func (app *application) healthcheckHandler(w http.ResponseWriter , r *http.Request){

	/*
	Fixed string format of the response body , this is useful for fast response for smaller body

	js := `{"status" : "available" , "environment" : "%s" , "version" : "%s" }`
	js = fmt.Sprintf(js , app.config.env , version)

	*/

	data := envelope{
		"status" : "available" ,
		"system_info":map[string]string{
			"environment" : app.config.env ,
			"version" : version ,
		},
	}

	err := app.writeJSON(w , http.StatusOK , data , nil)

	if err != nil {
		app.logger.Error(err.Error())
		http.Error(w , "The server encountered a problem and could not process your request" , http.StatusInternalServerError)
	}

}