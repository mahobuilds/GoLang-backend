package main

import (
	"fmt"
	"errors"
)

var ErrDeviceExists = errors.New("device already exists")

type DeviceNotFoundError struct {
	ID string
}

func (e *DeviceNotFoundError) Error() string {
	return fmt.Sprintf("device with ID: %s not found", e.ID)
}



