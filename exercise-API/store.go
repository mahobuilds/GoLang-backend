package main

import "sync"


type Store struct {
	workouts map[string]Workout
	mx sync.RWMutex
}

func NewStore() *Store {
	return &Store {
		workouts: make(map[string]Workout),
	}
}