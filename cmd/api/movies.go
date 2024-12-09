package main

import (
	"fmt"
	"net/http"
)


func (app *application) createMOvieHandler(w http.ResponseWriter, r *http.Request){
	fmt.Fprintln(w, "create a new movie")
}


func (app *application) showMovieHandler(w http.ResponseWriter, r *http.Request){
	id, err := app.readIDParams(r)
	if err != nil{
		http.NotFound(w, r)
		return
	}
	fmt.Fprintf(w, "show the details of movie %d\n", id)
}