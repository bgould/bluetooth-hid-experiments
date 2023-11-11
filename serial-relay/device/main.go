//go:build tinygo && circuitplay_bluefruit

package main

import (
	"machine"
	"time"

	"github.com/bgould/keyboard-firmware/hosts/multihost"
	"github.com/bgould/keyboard-firmware/hosts/serial"
	"github.com/bgould/keyboard-firmware/keyboard"
)

const _debug = true

var (
	pins    = []machine.Pin{machine.BUTTONA, machine.BUTTONB}
	layers  = CircuitPlaygroundDefaultKeymap()
	matrix  = keyboard.NewMatrix(1, 2, keyboard.RowReaderFunc(ReadRow))
	console = serial.DefaultConsole()
	client  = NewNUSClient()
)

func init() {

	// use the onboard LED as a status indicator
	machine.LED.Configure(machine.PinConfig{Mode: machine.PinOutput})
	machine.LED.Low()

	// configure pins for scanning matrix
	configurePins()

}

func main() {

	host := multihost.New(serial.New(machine.Serial), serial.New(client))
	board := keyboard.New(console, host, matrix, layers)
	board.SetDebug(_debug)

	time.Sleep(time.Second)

	for err := client.Init(); err != nil; err = client.Init() {
		println("Could not connect to NUS server:", err)
		time.Sleep(time.Second)
	}

	machine.LED.High()

	for {
		board.Task()
	}

}

func configurePins() {
	for _, pin := range pins {
		pin.Configure(machine.PinConfig{Mode: machine.PinInputPulldown})
		println("configured pin", pin, pin.Get())
	}
}

func ReadRow(rowIndex uint8) keyboard.Row {
	switch rowIndex {
	case 0:
		v := keyboard.Row(0)
		for i := range pins {
			if pins[i].Get() {
				v |= (1 << i)
			}
		}
		return v
	default:
		return 0
	}
}
