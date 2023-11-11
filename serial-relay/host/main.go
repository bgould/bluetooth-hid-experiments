//go:build tinygo && nrf

package main

import (
	"machine"
	"sync/atomic"

	"tinygo.org/x/bluetooth"
)

var (
	serviceUUID = bluetooth.ServiceUUIDNordicUART
	rxUUID      = bluetooth.CharacteristicUUIDUARTRX
	txUUID      = bluetooth.CharacteristicUUIDUARTTX

	buffer = machine.NewRingBuffer()
)

func main() {

	println("starting")
	adapter := bluetooth.DefaultAdapter
	must("enable BLE stack", adapter.Enable())

	adv := adapter.DefaultAdvertisement()
	must("config adv", adv.Configure(bluetooth.AdvertisementOptions{
		LocalName:    "NUS", // Nordic UART Service
		ServiceUUIDs: []bluetooth.UUID{serviceUUID},
	}))
	must("start adv", adv.Start())

	var rxChar bluetooth.Characteristic
	var txChar bluetooth.Characteristic

	must("add service", adapter.AddService(&bluetooth.Service{
		UUID: serviceUUID,
		Characteristics: []bluetooth.CharacteristicConfig{
			{
				Handle: &rxChar,
				UUID:   rxUUID,
				Flags:  bluetooth.CharacteristicWritePermission | bluetooth.CharacteristicWriteWithoutResponsePermission,
				WriteEvent: func(client bluetooth.Connection, offset int, value []byte) {
					// txChar.Write(value)
					for _, c := range value {
						buffer.Put(c)
					}
				},
			},
			{
				Handle: &txChar,
				UUID:   txUUID,
				Flags:  bluetooth.CharacteristicNotifyPermission | bluetooth.CharacteristicReadPermission,
			},
		},
	}))

	i := atomic.Int32{}
	for {
		if c, ok := buffer.Get(); ok {
			machine.Serial.WriteByte(c)
			v := i.Add(1)
			if v > 0xF {
				machine.Serial.WriteByte('\n')
				i.Store(0)
			}
		}
	}
}

func must(action string, err error) {
	if err != nil {
		panic("failed to " + action + ": " + err.Error())
	}
}
