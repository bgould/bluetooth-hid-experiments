package main

import (
	"io"

	"tinygo.org/x/bluetooth"
)

func NewNUSClient() *NUSClient {
	return &NUSClient{}
}

type NUSClient struct {
	rx bluetooth.DeviceCharacteristic
	tx bluetooth.DeviceCharacteristic
}

func (c *NUSClient) Init() error {

	adapter := bluetooth.DefaultAdapter

	// Enable BLE interface.
	err := adapter.Enable()
	if err != nil {
		println("could not enable the BLE stack:", err.Error())
		return err
	}

	// The address to connect to. Set during scanning and read afterwards.
	var foundDevice bluetooth.ScanResult

	// Scan for NUS peripheral.
	println("Scanning...")
	err = adapter.Scan(func(adapter *bluetooth.Adapter, result bluetooth.ScanResult) {
		if !result.AdvertisementPayload.HasServiceUUID(serviceUUID) {
			return
		}
		foundDevice = result

		// Stop the scan.
		err := adapter.StopScan()
		if err != nil {
			// Unlikely, but we can't recover from this.
			println("failed to stop the scan:", err.Error())
		}
	})
	if err != nil {
		println("could not start a scan:", err.Error())
		return err
	}

	// Found a device: print this event.
	if name := foundDevice.LocalName(); name == "" {
		print("Connecting to ", foundDevice.Address.String(), "...")
		println()
	} else {
		print("Connecting to ", name, " (", foundDevice.Address.String(), ")...")
		println()
	}

	// Found a NUS peripheral. Connect to it.
	device, err := adapter.Connect(foundDevice.Address, bluetooth.ConnectionParams{})
	if err != nil {
		println("Failed to connect:", err.Error())
		return err
	}

	// Connected. Look up the Nordic UART Service.
	println("Discovering service...")
	services, err := device.DiscoverServices([]bluetooth.UUID{serviceUUID})
	if err != nil {
		println("Failed to discover the Nordic UART Service:", err.Error())
		return err
	}
	service := services[0]

	// Get the two characteristics present in this service.
	chars, err := service.DiscoverCharacteristics([]bluetooth.UUID{rxUUID, txUUID})
	if err != nil {
		println("Failed to discover RX and TX characteristics:", err.Error())
		return err
	}
	c.rx = chars[0]
	c.tx = chars[1]

	// Enable notifications to receive incoming data.
	// err = tx.EnableNotifications(func(value []byte) {
	// 	for _, c := range value {
	// 		rawterm.Putchar(c)
	// 	}
	// })
	// if err != nil {
	// 	println("Failed to enable TX notifications:", err.Error())
	// 	return
	// }

	return nil
}

var _ io.Writer = (*NUSClient)(nil)

func (c *NUSClient) Write(data []byte) (int, error) {
	sendbuf := data // copy buffer
	// Reset the slice while keeping the buffer in place.
	data = data[:0]
	// Send the sendbuf after breaking it up in pieces.
	for len(sendbuf) != 0 {
		// Chop off up to 20 bytes from the sendbuf.
		partlen := 20
		if len(sendbuf) < 20 {
			partlen = len(sendbuf)
		}
		part := sendbuf[:partlen]
		sendbuf = sendbuf[partlen:]
		// This performs a "write command" aka "write without response".
		_, err := c.rx.WriteWithoutResponse(append(part))
		if err != nil {
			return 0, err
		}
	}
	return len(data), nil
}

var (
	serviceUUID = bluetooth.ServiceUUIDNordicUART
	rxUUID      = bluetooth.CharacteristicUUIDUARTRX
	txUUID      = bluetooth.CharacteristicUUIDUARTTX
)
