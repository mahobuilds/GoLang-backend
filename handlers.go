package main

import (
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"strconv"
)

type deviceGetter interface {
	GetDevice(id string) (Device, bool)
}

type allDevicesGetter interface {
	GetDevices() map[string]Device
}

type readingGetter interface {
	GetReadings(id string) ([]Reading, bool)
}

type deviceCreator interface {
	CreateDevice(device Device) error
}

type deviceReadingCreator interface {
	CreateDeviceReading(id string, reading Reading) error
}

type deviceReplacer interface {
	ReplaceDevice(device Device)
}

type deviceUpdater interface {
	UpdateDevice(id string, patch DevicePatch) (Device, bool)
}

type deviceDeleter interface {
	DeleteDevice(id string) bool
}

func createDeviceHandler(store deviceCreator) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		contentType := r.Header.Get("Content-Type")
		if contentType != "application/json" {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintln(w, "Mismatch of content type")
			return
		}

		var d Device
		err := json.NewDecoder(r.Body).Decode(&d)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintln(w, "bad server request: ", err)
			return
		}

		err = store.CreateDevice(d)
		if err != nil {
			w.WriteHeader(http.StatusConflict)
			fmt.Fprintf(w, "deivce with id: %s already exists", d.ID)
			return
		}
		json.NewEncoder(w).Encode(d)
	}
}

func createDeviceReading(store deviceReadingCreator) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		contentType := r.Header.Get("Content-Type")
		if contentType != "application/json" {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintln(w, "Mismatch of content type")
			return
		}

		var reading Reading

		err := json.NewDecoder(r.Body).Decode(&reading)

		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintln(w, "bad request:", err)
			return
		}

		id := r.PathValue("id")
		err = store.CreateDeviceReading(id, reading)

		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprintf(w, "Device with ID: %s couldn't be found", id)
			return
		}
		json.NewEncoder(w).Encode(reading)
	}
}

func getDeviceReading(store readingGetter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		id := r.PathValue("id")
		readings, exists := store.GetReadings(id)

		if !exists {
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprintf(w, "Device with ID: %s not found!", id)
			return
		}

		strFrom := r.URL.Query().Get("from")
		strTo := r.URL.Query().Get("to")

		var from float64
		to := math.MaxFloat64
		var ok error

		if strFrom != "" {
			from, ok = strconv.ParseFloat(strFrom, 64)
			if ok != nil {
				w.WriteHeader(http.StatusBadRequest)
				fmt.Fprintln(w, "bad request: ")
				return
			}
		}

		if strTo != "" {
			to, ok = strconv.ParseFloat(strTo, 64)
			if ok != nil {
				w.WriteHeader(http.StatusBadRequest)
				fmt.Fprintln(w, "bad request: ")
				return
			}
		}

		var included []Reading

		for _, reading := range readings {
			if reading.Timestamp >= from && reading.Timestamp <= to {
				included = append(included, reading)
			}
		}

		json.NewEncoder(w).Encode(included)

	}
}

func getAllDevices(store allDevicesGetter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var allDevices []Device

		for _, device := range store.GetDevices() {
			allDevices = append(allDevices, device)
		}
		json.NewEncoder(w).Encode(allDevices)
	}
}

func getDeviceData(store deviceGetter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		id := r.PathValue("id")

		device, exists := store.GetDevice(id)
		if !exists {
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprintf(w, "Device with ID: %s does not exist", id)
			return
		}

		json.NewEncoder(w).Encode(device)

	}
}

func getDeviceStats(store readingGetter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		id := r.PathValue("id")
		readings, exists := store.GetReadings(id)
		if !exists {
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprintf(w, "Device with ID: %s was not found!", id)
			return
		}

		if len(readings) == 0 {
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprintf(w, "Device with ID: %s has no readings!", id)
			return
		}

		strFrom := r.URL.Query().Get("from")
		strTo := r.URL.Query().Get("to")

		from := 0.0
		to := math.MaxFloat64
		var err error

		if strFrom != "" {
			from, err = strconv.ParseFloat(strFrom, 64)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				fmt.Fprintln(w, "bad request", err)
				return
			}
		}

		if strTo != "" {
			to, err = strconv.ParseFloat(strTo, 64)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				fmt.Fprintln(w, "bad request:", err)
				return
			}
		}

		type Response struct {
			Min     float64 `json:"min"`
			Max     float64 `json:"max"`
			Average float64 `json:"avg"`
		}

		var totalReadings []Reading
		if from > to {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintln(w, "invalid ranges")
			return
		}
		for _, reading := range readings {
			if reading.Timestamp >= from && reading.Timestamp <= to {
				totalReadings = append(totalReadings, reading)
			}
		}
		min, max, avg, ok := computeStats(totalReadings)
		if !ok {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		response := Response{Min: min, Max: max, Average: avg}
		json.NewEncoder(w).Encode(response)
	}
}

func replaceDevice(store deviceReplacer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		contentType := r.Header.Get("Content-Type")
		if contentType != "application/json" {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintln(w, "Mismatch of content type")
			return
		}

		var d Device

		err := json.NewDecoder(r.Body).Decode(&d)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintln(w, "bad request:", err)
			return
		}

		store.ReplaceDevice(d)
		json.NewEncoder(w).Encode(d)
	}
}

func updateDeviceData(store deviceUpdater) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		id := r.PathValue("id")

		contentType := r.Header.Get("Content-Type")
		if contentType != "application/json" {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintln(w, "Mismatch of content type")
			return
		}

		var patch DevicePatch

		err := json.NewDecoder(r.Body).Decode(&patch)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintln(w, "bad request:", err)
			return
		}

		device, exists := store.UpdateDevice(id, patch)
		if !exists {
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprintf(w, "Device with id: %s does not exist", id)
			return
		}
		json.NewEncoder(w).Encode(device)
	}
}

func deleteDevice(store deviceDeleter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")

		deleted := store.DeleteDevice(id)
		if !deleted {
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprintf(w, "Device with ID: %s does not exist", id)
			return
		}
		
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "request succeeded")
	}
}
