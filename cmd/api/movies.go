package main

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/LidoHon/LetsGOFurther-Greenlight.git/internal/data"
	"github.com/LidoHon/LetsGOFurther-Greenlight.git/internal/validator"
)


func (app *application) createMovieHandler(w http.ResponseWriter, r *http.Request){

	var input struct{
		Title 	string 			`json:"title"`
		Year 	int32 			`json:"year"`
		Runtime data.Runtime 	`json:"runtime"`
		Genres 	[]string		`json:"genres"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil{
		app.badRequestResponse(w, r, err)
		return
	}

	movie := &data.Movie{
		Title: input.Title,
		Year: input.Year,
		Runtime: input.Runtime,
		Genres: input.Genres,
	}
	v:= validator.New()

	
	if data.ValidateMovie(v, movie); !v.Valid(){
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = app.models.Movies.Insert(movie)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/v1/movies/%d", movie.ID))

	err = app.writeJSON(w, http.StatusCreated, envelop{"movie": movie}, headers)
	if err != nil{
		app.serverErrorResponse(w, r, err)
	}
	// fmt.Fprintf(w, "%+v\n", input)
}


func (app *application) showMovieHandler(w http.ResponseWriter, r *http.Request){

	// app.logger.Println("Received request for movie")
	id, err := app.readIDParams(r)
	if err != nil{
		// app.logger.Println("Error reading ID:", err)
		http.NotFound(w, r)
		return
	}
	// app.logger.Printf("Fetching movie with ID: %d\n", id)

	movie, err := app.models.Movies.Get(id)
	if err != nil{
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	// app.logger.Println("Movie fetched successfully. Sending response.")


	err = app.writeJSON(w, http.StatusOK,envelop{"movie": movie} , nil)
	if err !=nil{
		// app.logger.Println("Error writing JSON response:", err)
		app.serverErrorResponse(w,r,err)
	}
	
}

func(app *application) ListMoviesHandler( w http.ResponseWriter, r *http.Request){
	movies, err := app.models.Movies.GetAll()
	if err !=nil{
		switch{
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelop{"movies": movies}, nil)

	if err !=nil{
		app.serverErrorResponse(w, r, err)
	}
}


// update 

func (app *application) updateMovieHandler( w http.ResponseWriter, r *http.Request){

	id, err := app.readIDParams(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	movie, err := app.models.Movies.Get(id)
	if err != nil{
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w,r, err)
		}
		return
	}

	var input struct{
		Title 	*string 			`json:"title"`
		Year 	*int32 			`json:"year"`
		Runtime *data.Runtime 	`json:"runtime"`
		Genres 	[]string		`json:"genres"`
	}

	err = app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	if input.Title != nil{
		movie.Title = *input.Title
	}

	if input.Year != nil {
		movie.Year = *input.Year
	} 
	if input.Runtime != nil {
		movie.Runtime = *input.Runtime
	}
	if input.Genres !=nil {
		movie.Genres = input.Genres
	}

	v := validator.New()

	if data.ValidateMovie(v, movie); !v.Valid(){
		app.failedValidationResponse(w,r,v.Errors)
		return
	}

	err = app.models.Movies.Update(movie)
	if err != nil {
		switch{
		case errors.Is(err, data.ErrEditConflit):
			app.editConflictResponse(w, r)
		default:
			app.serverErrorResponse(w,r,err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelop{"movie": movie}, nil)

	if err !=nil {
		app.serverErrorResponse(w, r, err)
	}


}

func (app *application) DeleteMovieHandler(w http.ResponseWriter, r *http.Request){
	id, err := app.readIDParams(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	err = app.models.Movies.Delete(id)
	if err !=nil {
		switch{
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelop{"message": "movie sucessfully deleted"}, nil)
	if err !=nil{
		app.serverErrorResponse(w, r, err)
	}
}