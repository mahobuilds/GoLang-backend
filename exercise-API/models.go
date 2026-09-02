package main


type Workout struct {
	ID string `json:"id"`
	Name string `json:"name"`
	Exercises []string `json:"exercises"`
}