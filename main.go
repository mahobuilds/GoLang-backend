package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

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

	mux := http.NewServeMux()

	mux.HandleFunc("POST /devices", createDeviceHandler(store))

	mux.HandleFunc("POST /devices/{id}/readings", createDeviceReading(store))

	mux.HandleFunc("GET /devices", getAllDevices(store))

	mux.HandleFunc("GET /devices/{id}", getDevice(store))

	mux.HandleFunc("GET /devices/{id}/readings", getDeviceReading(store))

	mux.HandleFunc("GET /devices/{id}/stats", getDeviceStats(store))

	mux.HandleFunc("PUT /devices/{id}", replaceDevice(store))

	mux.HandleFunc("PATCH /devices/{id}", updateDeviceData(store))

	mux.HandleFunc("DELETE /devices/{id}", deleteDevice(store))

	mux.HandleFunc("/swagger/", httpSwagger.WrapHandler)

	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	signalCtx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	go func() {
		fmt.Println("Server running on: 8080")

		if err := server.ListenAndServe(); err != nil &&
			err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	<-signalCtx.Done()

	fmt.Println("Starting graceful shutdown...")

	shutdownCtx, cancel := context.WithTimeout(
		context.Background(),
		10*time.Second,
	)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	fmt.Println("Server stopped successfully")
}
