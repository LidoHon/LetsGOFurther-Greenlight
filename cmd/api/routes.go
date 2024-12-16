package main

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)


func (app *application) routes()http.Handler{

	router := httprouter.New()

	// error handlers
	router.NotFound = http.HandlerFunc(app.notFoundResponse)
	router.MethodNotAllowed = http.HandlerFunc(app.methodNotAllowedResponse)


	router.HandlerFunc(http.MethodGet, "/v1/healthcheck", app.healthcheckHandler)


	router.HandlerFunc(http.MethodPost, "/v1/movies", app.requirePermission("movies:write", app.createMovieHandler) )
	router.HandlerFunc(http.MethodGet, "/v1/movies/:id", app.requirePermission("movies:read", app.showMovieHandler) )
	router.HandlerFunc(http.MethodGet, "/v1/movies", app.requirePermission("movies:read", app.ListMoviesHandler) )
	router.HandlerFunc(http.MethodPatch, "/v1/movies/:id", app.requirePermission("movies:write", app.updateMovieHandler) )
	router.HandlerFunc(http.MethodDelete, "/v1/movies/:id", app.requirePermission("movies:write", app.DeleteMovieHandler) )

	// user routes.
	router.HandlerFunc(http.MethodPost, "/v1/users", app.registerUserHandler)
	router.HandlerFunc(http.MethodPost, "/v1/tokens/authentication", app.createAuthenticationTokenHandler)

	// activation
	router.HandlerFunc(http.MethodPut, "/v1/users/activated", app.activateUserHandler)

	return app.recoverPanic(app.rateLimit(app.authenticate(router)))
}