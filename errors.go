package main

import (
	"errors"
	"net/http"
	"encoding/json"
)

var ErrNoDevice = errors.New("deivce does not exist")
var ErrDeviceExists = errors.New("device already exists")

type ErrorResponse struct {
	Error string `json:"error"`
}

func writeError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-type", "application/json")
	w.WriteHeader(status)

	json.NewEncoder(w).Encode(ErrorResponse{
		Error: message,
	})

}