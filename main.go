package main

import "net/http"

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

	http.ListenAndServe(":8080", nil)
}
