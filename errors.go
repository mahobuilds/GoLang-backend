package main

import "errors"

var ErrNoDevice = errors.New("deivce does not exist")
var ErrDeviceExists = errors.New("device already exists")