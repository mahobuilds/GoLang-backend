package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"math"
	"strconv"
)

func createDeviceHandler(store *Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		store.mx.Lock()
		defer store.mx.Unlock()

		var d Device
		err := json.NewDecoder(r.Body).Decode(&d)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintln(w, "bad server request: ", err)
			return
		}

		_, exists := store.devices[d.ID]
		if exists {
			w.WriteHeader(http.StatusConflict)
			fmt.Fprintf(w, "A device with ID: %s already exists!", d.ID)
			return
		}

		store.devices[d.ID] = d
		json.NewEncoder(w).Encode(d)

	}
}



func createDeviceReading(store *Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		store.mx.Lock()
		defer store.mx.Unlock()

		var reading Reading

		err := json.NewDecoder(r.Body).Decode(&reading)

		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintln(w, "bad request:", err)
			return
		}

		id := r.PathValue("id")
		_, exists := store.devices[id]

		if !exists {
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprintf(w, "Device with ID: %s couldn't be found", id)
			return
		}

		store.readings[id] = append(store.readings[id], reading)
		json.NewEncoder(w).Encode(reading)
	}
}



func getDeviceReading(store *Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		store.mx.RLock()
		defer store.mx.RUnlock()

		id := r.PathValue("id")
		_, exists := store.devices[id]

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

		for _, reading := range store.readings[id] {
			if reading.Timestamp >= from && reading.Timestamp <= to {
				included = append(included, reading)
			}
		}

		json.NewEncoder(w).Encode(included)

	}
}



func getDeviceStats(store *Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		store.mx.RLock()
		defer store.mx.RUnlock()

		id := r.PathValue("id")
		_, exists := store.devices[id]
		if !exists {
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprintf(w, "Device with ID: %s was not found!", id)
			return
		}

		if len(store.readings[id]) == 0 {
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprintf(w, "Device with ID: %s has no readings!", id)
			return
		}

		min, max, avg, ok := computeStats(store.readings[id])
		if !ok {
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprintf(w, "Device with ID: %s has no readings!", id)
			return
		}

		response := struct {
			Min     float64 `json:"min"`
			Max     float64 `json:"max"`
			Average float64 `json:"avg"`
		}{Min: min, Max: max, Average: avg}

		json.NewEncoder(w).Encode(response)
	}
}



