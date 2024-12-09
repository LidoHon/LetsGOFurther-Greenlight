package main

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/julienschmidt/httprouter"
)


func (app *application) readIDParams(r *http.Request)(int64, error){
	parmas := httprouter.ParamsFromContext(r.Context())
	id, err := strconv.ParseInt(parmas.ByName("id"), 10, 64)
	if err !=nil || id < 1{
		return 0, errors.New("invalid id parameter")
	}
	return id, nil
}