package main 

import "net/http"


func main() {

	store := NewStore()

	http.HandleFunc("POST /workouts", addNewWorkout(store))

	http.HandleFunc("GET /workouts", getAllWorkouts(store))

	http.HandleFunc("GET /workouts/{id}", getWorkout(store))

	http.HandleFunc("DELETE /workouts/{id}", deleteWorkout(store))

	http.ListenAndServe(":8080", nil)

}