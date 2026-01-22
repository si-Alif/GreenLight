package main

import (
	"net/http"
	// "time"
)

func (app *application) healthcheckHandler(w http.ResponseWriter, r *http.Request) {

	/*
		Fixed string format of the response body , this is useful for fast response for smaller body

		js := `{"status" : "available" , "environment" : "%s" , "version" : "%s" }`
		js = fmt.Sprintf(js , app.config.env , version)

	*/

	data := envelope{
		"status": "available",
		"system_info": map[string]string{
			"environment": app.config.env,
			"version":     version,
		},
	}

	// time.Sleep(5 * time.Second)

	err := app.writeJSON(w, http.StatusOK, data, nil)

	if err != nil {
		app.serverErrorResponse(w, r, err)
	}

}
