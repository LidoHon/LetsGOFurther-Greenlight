package main

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/LidoHon/LetsGOFurther-Greenlight.git/internal/data"
	"github.com/LidoHon/LetsGOFurther-Greenlight.git/internal/validator"
	"github.com/pascaldekloe/jwt"
)

func (app *application) createAuthenticationTokenHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	v := validator.New()

	data.ValidateEmail(v, input.Email)
	data.ValidPasswordPlaintext(v, input.Password)

	if !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	user, err := app.models.Users.GetByEmail(input.Email)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.invalidCredentialResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	match, err := user.Password.Matches(input.Password)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	if !match {
		app.invalidCredentialResponse(w, r)
		return
	}

	var claims jwt.Claims
	claims.Subject = strconv.FormatInt(user.ID, 10)
	claims.Issued = jwt.NewNumericTime(time.Now())
	claims.NotBefore = jwt.NewNumericTime(time.Now())
	claims.Expires = jwt.NewNumericTime(time.Now().Add(24 * time.Hour))
	claims.Issuer = "liduhon3@gmail.com"
	claims.Audiences = []string{"liduhon3@gmail.com"}

	jwtBytes, err := claims.HMACSign(jwt.HS256, []byte(app.config.jwt.secret))
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	err = app.writeJSON(w, http.StatusCreated, envelope{"authentication_token": string(jwtBytes)}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}

	// token, err := app.models.Tokens.New(user.ID, 24*time.Hour, data.ScopeAuthentication)
	// if err != nil {
	// 	app.serverErrorResponse(w, r, err)
	// 	return
	// }

	// err = app.writeJSON(w, http.StatusCreated, envelope{"authentication_token": token}, nil)
	// if err != nil {
	// 	app.serverErrorResponse(w, r, err)
	// }

}
