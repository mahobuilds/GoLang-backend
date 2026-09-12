package main

import (
	"sync"
	"context"
)

type Store struct {
	devices  map[string]Device
	readings map[string][]Reading
	mx       sync.RWMutex
}

func (store *Store) GetDevice(ctx context.Context, id string) (Device, error) {
	select {
	case <- ctx.Done():
		return Device{}, ctx.Err()
	default:
	}

	store.mx.RLock()
	defer store.mx.RUnlock()

	device, exists := store.devices[id]
	if !exists {
		return  Device{}, ErrNoDevice
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

func (store *Store) GetReadings(id string) ([]Reading, error) {
	store.mx.RLock()
	defer store.mx.RUnlock()

	readings, exists := store.readings[id]
	if !exists {
		return []Reading{}, ErrNoDevice
	}

	return readings, nil
}

func (store *Store) CreateDevice(device Device) error {
	store.mx.Lock()
	defer store.mx.Unlock()

	_, exists := store.devices[device.ID]
	if exists {
		return ErrDeviceExists
	}

	store.devices[device.ID] = device
	return nil
}

func (store *Store) CreateDeviceReading(id string, reading Reading) error {
	store.mx.Lock()
	defer store.mx.Unlock()

	_, exists := store.devices[id]
	if !exists {
		return ErrNoDevice
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

func (store *Store) UpdateDevice(id string, patch DevicePatch) (Device, error) {
	store.mx.Lock()
	defer store.mx.Unlock()

	device, exists := store.devices[id]
	if !exists {
		return Device{}, ErrNoDevice
	}

	if patch.Name != nil {
		device.Name = *patch.Name
	}
	if patch.Type != nil {
		device.Type = *patch.Type
	}

	store.devices[device.ID] = device
	return device, nil
}

func (store *Store) DeleteDevice(id string) error {
	store.mx.Lock()
	defer store.mx.Unlock()
	
	_, exists := store.devices[id]
	if !exists {
		return ErrNoDevice
	}

	delete(store.devices, id)
	delete(store.readings, id)
	return nil
}

func NewStore() *Store {
	return &Store{
		devices:  make(map[string]Device),
		readings: make(map[string][]Reading),
	}
}
