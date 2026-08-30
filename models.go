package main

type Device struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"`
}

type Reading struct {
	Value     float64 `json:"value"`
	Timestamp float64 `json:"timestamp"`
}
