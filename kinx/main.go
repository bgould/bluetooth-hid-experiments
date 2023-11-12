//go:build tinygo && feather_nrf52840

package main

import (
	"machine"
	"runtime"
	"time"

	"github.com/bgould/keyboard-firmware/hosts/multihost"
	"github.com/bgould/keyboard-firmware/hosts/serial"
	"github.com/bgould/keyboard-firmware/keyboard"
)

const (
	_debug = true
)

var (
	ble    = &bleHost{}
	host   = configureHost()
	keymap = Keymap()
	board  = keyboard.New(console, host, matrix, keymap)
)

func main() {

	time.Sleep(time.Second)

	if _debug {
		time.Sleep(3 * time.Second)
		println("initializing hardware")
	}

	configureMatrix()

	println("Configuring BLE")
	for configured := false; !configured; {
		if err := configureBLE(); err != nil {
			println("error configuring BLE", err.Error())
			configured = false
			time.Sleep(time.Second)
		}
		configured = true
	}
	println("BLE configured successfully")

	if _debug {
		println("starting task loop")
	}
	board.SetDebug(_debug)

	go bootBlink()
	go deviceLoop()

	for {
		runtime.Gosched()
		// time.Sleep(1 * time.Second)
	}
}

func deviceLoop() {
	for {
		board.Task()
		runtime.Gosched()
	}
}

func configureBLE() error {
	if err := ble.Enable(); err != nil {
		return err
	}
	if err := ble.Connect(); err != nil {
		if _debug {
			println("failed to establish LESC connection", err.Error())
		}
		return err
	}
	if _debug {
		println("established LESC connection")
	}
	ble.RegisterHID()
	if _debug {
		println("registered HID")
	}
	return nil
}

func configureHost() keyboard.Host {
	// if !_debug {
	// 	return usbhid.New()
	// }
	return multihost.New(ble, serial.New(machine.Serial))
}
