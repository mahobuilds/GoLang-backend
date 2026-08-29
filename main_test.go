package main

import (
	"sync"
	"testing"
)

func TestComputeStats(t *testing.T) {
	readings := []Reading{
		{Value: 10, Timestamp: 1},
		{Value: 20, Timestamp: 2},
		{Value: 30, Timestamp: 3},
	}

	min, max, avg, ok := computeStats(readings)

	if !ok {
		t.Fatalf("expected ok=true for non-empty readings")
	}
	if min != 10 {
		t.Errorf("expected min=10, got %v", min)
	}
	if max != 30 {
		t.Errorf("expected max=30, got %v", max)
	}
	if avg != 20 {
		t.Errorf("expected avg=20, got %v", avg)
	}
}

func TestComputeStatsEmpty(t *testing.T) {
	var readings []Reading

	min, max, avg, ok := computeStats(readings)

	if ok {
		t.Fatalf("expected ok=false for empty readings")
	}
	if min != 0 || max != 0 || avg != 0 {
		t.Errorf("expected zero values on empty input, got min=%v max=%v avg=%v", min, max, avg)
	}
}

func TestConcurrentReadingsWrite(t *testing.T) {
	var mx sync.RWMutex
	readings := make(map[string][]Reading)
	id := "device-1"

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(val float64) {
			defer wg.Done()
			mx.Lock()
			readings[id] = append(readings[id], Reading{Value: val, Timestamp: val})
			mx.Unlock()
		}(float64(i))
	}
	wg.Wait()

	if len(readings[id]) != 100 {
		t.Errorf("expected 100 readings, got %d", len(readings[id]))
	}
}
