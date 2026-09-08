package main

import (
	"fmt"
	"sync"
)

type Store struct {
	devices  map[string]Device
	readings map[string][]Reading
	mx       sync.RWMutex
}

func (store *Store) GetDevice(id string) (Device, error) {
	store.mx.RLock()
	defer store.mx.RUnlock()

	device, exists := store.devices[id]
	if !exists {
		return Device{}, &DeviceNotFoundError{
			ID: id,
		}
	}

	return device, nil
}

func (store *Store) GetDevices() map[string]Device {
	store.mx.RLock()
	defer store.mx.RUnlock()

	devices := make(map[string]Device, len(store.devices))
	for id, device := range store.devices {
		devices[id] = device
	}

	return devices
}

func (store *Store) GetReadings(id string) ([]Reading, bool) {
	store.mx.RLock()
	defer store.mx.RUnlock()

	readings, exists := store.readings[id]
	return readings, exists
}

func (store *Store) CreateDevice(device Device) error {
	store.mx.Lock()
	defer store.mx.Unlock()

	_, exists := store.devices[device.ID]
	if exists {
		return fmt.Errorf("creating device %s: %w", device.ID, ErrDeviceExists)
	}

	store.devices[device.ID] = device
	return nil
}

func (store *Store) CreateDeviceReading(id string, reading Reading) error {
	store.mx.Lock()
	defer store.mx.Unlock()

	_, exists := store.devices[id]
	if !exists {
		return &DeviceNotFoundError{ID : id}
	}

	store.readings[id] = append(store.readings[id], reading)
	return nil
}

func (store *Store) ReplaceDevice(device Device) {
	store.mx.Lock()
	defer store.mx.Unlock()

	_, exists := store.devices[device.ID]
	if exists {
		newDevice := store.devices[device.ID]

		newDevice.ID = device.ID
		newDevice.Name = device.Name
		newDevice.Type = device.Type

		store.devices[device.ID] = newDevice
	} else {
		store.devices[device.ID] = device
	}

	
}

func (store *Store) UpdateDevice(id string, patch DevicePatch) (Device, bool) {
	store.mx.Lock()
	defer store.mx.Unlock()

	device, exists := store.devices[id]
	if !exists {
		return Device{}, false
	}

	if patch.Name != nil {
		device.Name = *patch.Name
	}
	if patch.Type != nil {
		device.Type = *patch.Type
	}

	store.devices[device.ID] = device
	return device, true
}

func (store *Store) DeleteDevice(id string) bool {
	store.mx.Lock()
	defer store.mx.Unlock()
	
	_, exists := store.devices[id]
	if !exists {
		return false
	}

	delete(store.devices, id)
	delete(store.readings, id)
	return true
}

func NewStore() *Store {
	return &Store{
		devices:  make(map[string]Device),
		readings: make(map[string][]Reading),
	}
}
