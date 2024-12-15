package main

import (
	"errors"
	"net/http"
	"time"

	"github.com/LidoHon/LetsGOFurther-Greenlight.git/internal/data"
	"github.com/LidoHon/LetsGOFurther-Greenlight.git/internal/validator"
)



func(app *application) registerUserHandler(w http.ResponseWriter, r *http.Request){

	var input struct{
		Name string `json:"name"`
		Email string `json:"email"`
		Password string `json:"password"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	user := &data.User{
		Name: input.Name,
		Email: input.Email,
		Activated: false,

	}

	err = user.Password.Set(input.Password)
	if err !=nil {
		app.serverErrorResponse(w,r, err)
		return
	}

	v := validator.New()
	if data.ValidateUser(v, user); !v.Valid(){
		app.failedValidationResponse(w,r,v.Errors)
		return
	}

	err = app.models.Users.Insert(user)
	if err !=nil {
		switch{
		case errors.Is(err, data.ErrDuplicateEmail):
			v.AddError("email", "a user with this email alrady exists")
			app.failedValidationResponse(w,r, v.Errors)
		default:
			app.serverErrorResponse(w,r,err)

		}
		return
	}

	token, err := app.models.Tokens.New(user.ID, 3*24*time.Hour, data.ScopeActivation)
	if err !=nil {
		app.serverErrorResponse(w,r,err)
	}

	app.background(func() {
		data := map[string]interface{}{
			"activationToken": token.PlainText,
			"userID": user.ID,
			"name": user.Name,
		}
		app.logger.PrintInfo("User name for email: " + user.Name, nil)
		err = app.mailer.Send(user.Email, "user_welcome.tmpl", data)
		if err != nil {
			app.logger.PrintError(err, nil)
		}

	})

		
	

	err = app.writeJSON(w, http.StatusAccepted, envelope{"user": user}, nil)
	if err != nil {
		app.serverErrorResponse(w,r,err)
	}

}

func (app *application) activateUserHandler(w http.ResponseWriter, r *http.Request){
	var input struct{
		TokenPlaintext string `json:"token"`
	}

	err := app.readJSON(w,r, &input)
	if err !=nil {
		app.badRequestResponse(w,r, err)
		return
	}

	v := validator.New()

	if data.ValidateTokenPlaintext(v, input.TokenPlaintext);!v.Valid(){
		app.failedValidationResponse(w,r, v.Errors)
		return
	}

	user, err := app.models.Users.GetForToken(data.ScopeActivation, input.TokenPlaintext)
	if err !=nil {
		switch{
		case errors.Is(err, data.ErrRecordNotFound):
			v.AddError("token","invalid or expired activation token")
			app.failedValidationResponse(w,r, v.Errors)
		default:
			app.serverErrorResponse(w,r,err)
		}
		return
	}

	user.Activated = true

	err = app.models.Users.Update(user)
	if err !=nil {
		switch{
		case errors.Is(err, data.ErrEditConflit):
			app.editConflictResponse(w,r)
		default:
			app.serverErrorResponse(w,r,err)
		}
		return
	}

	err = app.models.Tokens.DeleteAllForUser(data.ScopeActivation, user.ID)
	if err !=nil {
		app.serverErrorResponse(w,r,err)
		return
	}
	err = app.writeJSON(w, http.StatusOK, envelope{"user":user}, nil)
	if err != nil {
		app.serverErrorResponse(w,r, err)
	}
}