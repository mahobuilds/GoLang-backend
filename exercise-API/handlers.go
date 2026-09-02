package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func addNewWorkout(store *Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		store.mx.Lock()
		defer store.mx.Unlock()

		contentType := r.Header.Get("Content-Type")
		if contentType != "application/json" {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintln(w, "Mismatch of content type")
			return
		}

		var work Workout

		err := json.NewDecoder(r.Body).Decode(&work)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintln(w, "bad request:", err)
			return
		}

		_, exists := store.workouts[work.ID]
		if exists {
			w.WriteHeader(http.StatusConflict)
			fmt.Fprintf(w, "Workout with ID: %s already exists", work.ID)
			return
		}

		store.workouts[work.ID] = work
		json.NewEncoder(w).Encode(work)
	}
}

func getAllWorkouts(store *Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		store.mx.RLock()
		defer store.mx.RUnlock()

		var workouts []Workout

		for _, work := range store.workouts {
			workouts = append(workouts, work)
		}

		json.NewEncoder(w).Encode(workouts)
	}
}

func getWorkout(store *Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		store.mx.RLock()
		defer store.mx.RUnlock()

		id := r.PathValue("id")
		_, exists := store.workouts[id]
		if !exists {
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprintf(w, "workout with ID: %s not found", id)
			return
		}
		json.NewEncoder(w).Encode(store.workouts[id])
	}
}

func deleteWorkout(store *Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		store.mx.RLock()
		defer store.mx.RUnlock()

		id := r.PathValue("id")

		_, exists := store.workouts[id]
		if !exists {
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprintf(w, "workout with ID: %s does not exist", id)
			return
		}

		delete(store.workouts, id)
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "Workout deleted successfully")
	}
}
