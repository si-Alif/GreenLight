package main

import (
	"errors"
	"net/http"
	"time"

	"greenlight.si-Alif.net/internal/data"
	"greenlight.si-Alif.net/internal/validator"
)

func (app *application) registerUserHandler(w http.ResponseWriter , r *http.Request){
	var input struct{
		Name string `json:"name"`
		Email string `json:"email"`
		Password string `json:"password"`
	}
	err := app.readJSON(w , r , &input)

	if err != nil {
		app.badRequestResponse(w , r , err)
		return
	}

	user := &data.User{
		Name: input.Name,
		Email: input.Email,
		Activated: false,
	}

	err = user.Password.Set(input.Password)
	if err != nil {
		app.serverErrorResponse(w , r , err)
		return
	}

	v := validator.New()

	if data.ValidateUser(v , user);!v.Valid(){
		app.failedValidationResponse(w , r , v.Errors)
		return
	}

	err = app.models.Users.Insert(user)
	if err != nil {
		switch{
			case errors.Is(err , data.ErrDuplicateEmail):
				v.AddError("email" , "a user with this email address already exists")
				app.failedValidationResponse(w , r , v.Errors)
			default :
				app.serverErrorResponse(w , r , err)
		}
		return
	}

	// generate a activation token from the user
	token , err := app.models.Tokens.New(user.ID , 24*time.Hour , data.ScopeActivation)
	if err != nil{
		app.serverErrorResponse(w , r , err)
		return
	}

	// send email using the mailer.Send() method before returning http response
	templateData := struct {
		User             *data.User
		RegistrationDate string
		CurrentYear      int
		ActivationToken  string
		UserID           int64
	}{
		User:             user,
		RegistrationDate: user.CreatedAt.Format("January 2, 2006"),
		CurrentYear:      time.Now().Year(),
		ActivationToken:  token.PlainText,
		UserID:           user.ID,
	}

	// place the email sending functionality in a background goroutine
	app.background(func(){
		err := app.mailer.Send(user.Email , "user_welcome.tmpl.html" , templateData)
		if err != nil{
			app.logger.Error(err.Error()) // not sending app.serverErrorResponse() cz by the time we encounter any error from this , the client would've already sent the http response and even if we use it , it would send another http response resulting in a error
		}
	})

	// err = app.writeJSON(w , http.StatusCreated , envelope{"user":user} , nil) --> rather than this we'll use http.StatusAccepted code to make the user realize that their has been accepted for processing
	err = app.writeJSON(w , http.StatusAccepted , envelope{"user":user} , nil)

	if err != nil{
		app.serverErrorResponse(w , r , err)
	}
}
