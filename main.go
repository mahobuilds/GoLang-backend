package main

import (
	"net/http"

	_ "iotmer-case/docs"

	httpSwagger "github.com/swaggo/http-swagger"
)

// @title IoT Device API
// @version 1.0
// @description A REST API for managing IoT devices and their readings

// @host localhost:8080
// @BasePath /

func main() {

	store := NewStore()

	http.HandleFunc("POST /devices", createDeviceHandler(store))

	http.HandleFunc("POST /devices/{id}/readings", createDeviceReading(store))

	http.HandleFunc("GET /devices", getAllDevices(store))

	http.HandleFunc("GET /devices/{id}", getDeviceData(store))

	http.HandleFunc("GET /devices/{id}/readings", getDeviceReading(store))

	http.HandleFunc("GET /devices/{id}/stats", getDeviceStats(store))

	http.HandleFunc("PUT /devices/{id}", replaceDevice(store))

	http.HandleFunc("PATCH /devices/{id}", updateDeviceData(store))

	http.HandleFunc("DELETE /devices/{id}", deleteDevice(store))

	http.HandleFunc("/swagger/", httpSwagger.WrapHandler)

	http.ListenAndServe(":8080", nil)
}
