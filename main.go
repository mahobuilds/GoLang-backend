package main

import "net/http"

func main() {

	store := NewStore()

	http.HandleFunc("POST /devices", createDeviceHandler(store))

	http.HandleFunc("POST /devices/{id}/readings", createDeviceReading(store))

	http.HandleFunc("GET /devices/{id}/readings", getDeviceReading(store))

	http.HandleFunc("GET /devices/{id}/stats", getDeviceStats(store))

	http.ListenAndServe(":8080", nil)
}
