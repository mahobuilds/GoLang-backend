package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/http"
	"strconv"
)

type StatsResponse struct {
	Min     float64 `json:"min"`
	Max     float64 `json:"max"`
	Average float64 `json:"avg"`
}

type deviceGetter interface {
	GetDevice(id string) (Device, error)
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

// creatDeviceHandler godoc
// @Summary Create a new IoT device
// @Description Create a new deivce and save its parameters
// @Tags devices
// @Accept json
// @Produce json
// @Param device body Device true "Device to create"
// @Success 200 {object} Device
// @Failure 400 {string} string
// @Failure 409 {string} string
// @Router /devices [post]
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
			if errors.Is(err, ErrDeviceExists) {
				w.WriteHeader(http.StatusConflict)
				fmt.Fprintf(w, "device with id: %s already exists", d.ID)
				return
			}
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintln(w, "internal server error")
			return
		}
		json.NewEncoder(w).Encode(d)
	}
}

// createDeviceReading godoc
// @Summary Add a device reading
// @Description Create a new device reading and store it
// @Tags readings
// @Accept json
// @Produce json
// @Param id path string true "Device ID"
// @Param reading body Reading true "Reading to be added"
// @Success 200 {object} Reading
// @Failure 400 {string} string
// @Failure 404 {string} string
// @Router /devices/{id}/readings [post]
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
			var notFoundErr *DeviceNotFoundError
			if errors.As(err, &notFoundErr) {
				w.WriteHeader(http.StatusNotFound)
				fmt.Fprintf(w, "Device with ID: %s does not exist", notFoundErr.ID)
				return
			}
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintln(w, "internal server error")
			return
		}
		json.NewEncoder(w).Encode(reading)
	}
}

// getDeviceReading godoc
// @Summary Get a device's reading
// @Description Get a device's readings by its given id
// @Tags readings
// @Produce json
// @Param id path string true "Device ID"
// @Param from query number false "Start timestamp"
// @Param to query number false "End timestamp"
// @Success 200 {object} Reading
// @Failure 404 {string} string
// @Failure 400 {string} string
// @Router /devices/{id}/readings [get]
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

// getAllDevices
// @Summary Get all devices
// @Description Get all devices stored
// @Tags devices
// @Produce json
// @Success 200 {array} Device
// @Failure 404 {string} string
// @Router /devices [get]
func getAllDevices(store allDevicesGetter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var allDevices []Device

		for _, device := range store.GetDevices() {
			allDevices = append(allDevices, device)
		}
		json.NewEncoder(w).Encode(allDevices)
	}
}

// getDeviceData godoc
// @Summary Get a device
// @Description Get a single device by its ID
// @Tags devices
// @Produce json
// @Param id path string true "Device ID"
// @Success 200 {object} Device
// @Failure 404 {string} string
// @Router /devices/{id} [get]
func getDeviceData(store deviceGetter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		id := r.PathValue("id")

		device, err := store.GetDevice(id)
		if err != nil {
			var notFoundErr *DeviceNotFoundError
			if errors.As(err, &notFoundErr) {
				w.WriteHeader(http.StatusNotFound)
				fmt.Fprintf(w, "Device with ID: %s does not exist", notFoundErr.ID)
				return
			}
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintln(w, "internal server error")
			return
		}

		json.NewEncoder(w).Encode(device)

	}
}

// getDeviceStats godoc
// @Summary Get a device's statistics
// @Description Calculate the minimum, maximum, and average values for a device's readings
// @Tags stats
// @Produce json
// @Param id path string true "Device ID"
// @Param from query number false "Start timestamp"
// @Param to query number false "End timestamp"
// @Success 200 {object} StatsResponse
// @Failure 404 {string} string
// @Failure 400 {string} string
// @Failure 204 {string} string
// @Router /devices/{id}/reading/stats [get]
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

		response := StatsResponse{Min: min, Max: max, Average: avg}
		json.NewEncoder(w).Encode(response)
	}
}

// replaceDevice godoc
// @Summary Replace a device
// @Description If the device exists, replace it with the new one. If not, create a new device
// @Tags devices
// @Accept json
// @Produce json
// @Param id path string true "Device ID"
// @Param device body Device true "Device to be replaced"
// @Success 200 {object} Device
// @Failure 400 {string} string
// @Failure 404 {string} string
// @Router /devices/{id} [PUT]
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

// updateDeviceData godoc
// @Summary Update device data
// @Description Update the data of an already existing device
// @Tags devices
// @Accept json
// @Produce json
// @Param id path string true "Device ID"
// @Param devicePatch body DevicePatch true "Data to be updated"
// @Success 200 {Object} Device
// @Failure 400 {string} string
// @Failure 404 {stirng} string
// @Router /devices/{id} [PATCH]
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

// deleteDevice godoc
// @Summary Delete a device
// @Description Delete a device and its readings
// @Tags devices
// @Param id path string true "Device ID"
// @Success 200
// @Failure 404 {string} string
// @Router /devices/{id} [DELETE]
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
