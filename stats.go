package main

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