package main

import (
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"sync"
)

type Device struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"`
}

type Reading struct {
	Value     float64 `json:"value"`
	Timestamp float64 `json:"timestamp"`
}

func computeStats(readings []Reading) (min, max, avg float64, ok bool) {
	if len(readings) == 0 {
		return 0, 0, 0, false
	}

	min = readings[0].Value
	max = readings[0].Value
	sum := 0.0

	for _, reading := range readings {
		if min > reading.Value {
			min = reading.Value
		}
		if max < reading.Value {
			max = reading.Value
		}
		sum += reading.Value
	}

	avg = sum / float64(len(readings))
	return min, max, avg, true
}

func main() {

	devices := make(map[string]Device)
	readings := make(map[string][]Reading)
	var mx sync.RWMutex

	http.HandleFunc("POST /devices", func(w http.ResponseWriter, r *http.Request) {
		mx.Lock()
		defer mx.Unlock()

		var d Device
		err := json.NewDecoder(r.Body).Decode(&d)

		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintln(w, "bad server request:", err)
			return
		}

		_, exists := devices[d.ID]
		if exists {
			w.WriteHeader(http.StatusConflict)
			fmt.Fprintf(w, "A device with ID: %s already exists!", d.ID)
			return
		}

		devices[d.ID] = d
		json.NewEncoder(w).Encode(d)
	})

	http.HandleFunc("POST /devices/{id}/readings", func(w http.ResponseWriter, r *http.Request) {
		mx.Lock()
		defer mx.Unlock()

		var reading Reading

		err := json.NewDecoder(r.Body).Decode(&reading)

		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintln(w, "bad request:", err)
			return
		}

		id := r.PathValue("id")
		_, exists := devices[id]

		if !exists {
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprintf(w, "Device with ID: %s couldn't be found", id)
			return
		}

		readings[id] = append(readings[id], reading)
		json.NewEncoder(w).Encode(reading)
	})

	http.HandleFunc("GET /devices/{id}/readings", func(w http.ResponseWriter, r *http.Request) {
		mx.RLock()
		defer mx.RUnlock()

		id := r.PathValue("id")
		_, exists := devices[id]

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

		for _, reading := range readings[id] {
			if reading.Timestamp >= from && reading.Timestamp <= to {
				included = append(included, reading)
			}
		}

		json.NewEncoder(w).Encode(included)

	})

	http.HandleFunc("GET /devices/{id}/stats", func(w http.ResponseWriter, r *http.Request) {
		mx.RLock()
		defer mx.RUnlock()

		id := r.PathValue("id")
		_, exists := devices[id]
		if !exists {
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprintf(w, "Device with ID: %s was not found!", id)
			return
		}

		if len(readings[id]) == 0 {
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprintf(w, "Device with ID: %s has no readings!", id)
			return
		}

		min, max, avg, ok := computeStats(readings[id])
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

	})

	http.ListenAndServe(":8080", nil)
}
