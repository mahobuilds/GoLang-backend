package main

import "sync"

type Store struct {
	devices map[string]Device
	readings map[string][]Reading
	mx sync.RWMutex
}

func NewStore() *Store {
	return &Store {
		devices: make(map[string]Device),
		readings: make(map[string][]Reading),
	}
}

