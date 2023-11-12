//go:build tinygo && nrf52840

package main

import (
	"github.com/bgould/keyboard-firmware/keyboard"
	"tinygo.org/x/bluetooth"
)

type bleHost struct {
	reportIn bluetooth.Characteristic

	keybuf   [9]byte
	conbuf   [9]byte
	mousebuf [5]byte
}

var _ keyboard.Host = (*bleHost)(nil)

func (ble *bleHost) Enable() error {
	adapter := bluetooth.DefaultAdapter
	err := adapter.Enable()
	if err != nil {
		return err
	}
	return nil
}

func (ble *bleHost) Connect() error {
	adapter := bluetooth.DefaultAdapter

	// peerPKey := make([]byte, 0, 64)
	// privLesc, err := ecdh.P256().GenerateKey(rand.Reader)
	// if err != nil {
	// 	return err
	// }
	// lescChan := make(chan struct{})
	bluetooth.SetSecParamsBonding()
	//bluetooth.SetSecParamsLesc()
	bluetooth.SetSecCapabilities(bluetooth.NoneGapIOCapability)
	// time.Sleep(4 * time.Second)
	// println("getting own pub key")
	// var key []byte

	// pk := privLesc.PublicKey().Bytes()
	// pubKey := swapEndinan(pk[1:])
	//bluetooth.SetLesPublicKey(swapEndinan(privLesc.PublicKey().Bytes()[1:]))
	// pubKey = nil
	//println(" key is set")

	// println("register lesc callback")
	// adapter.SetLescRequestHandler(
	// 	func(pubKey []byte) {
	// 		peerPKey = pubKey
	// 		close(lescChan)
	// 	},
	// )

	println("def adv")
	adv := adapter.DefaultAdvertisement()
	println("adv config")
	adv.Configure(bluetooth.AdvertisementOptions{
		LocalName: "tinygo-kinadv2",
		ServiceUUIDs: []bluetooth.UUID{
			bluetooth.ServiceUUIDDeviceInformation,
			bluetooth.ServiceUUIDBattery,
			bluetooth.ServiceUUIDHumanInterfaceDevice,
		},
	})
	println("adv start")
	return adv.Start()
}

func (ble *bleHost) RegisterHID() {
	adapter := bluetooth.DefaultAdapter
	adapter.AddService(&bluetooth.Service{
		UUID: bluetooth.ServiceUUIDDeviceInformation,
		Characteristics: []bluetooth.CharacteristicConfig{
			{
				UUID:  bluetooth.CharacteristicUUIDManufacturerNameString,
				Flags: bluetooth.CharacteristicReadPermission,
				Value: []byte("Kinesis Computer Ergonomics"),
			},
			{
				UUID:  bluetooth.CharacteristicUUIDModelNumberString,
				Flags: bluetooth.CharacteristicReadPermission,
				Value: []byte("Advantage2"),
			},
			{
				UUID:  bluetooth.CharacteristicUUIDPnPID,
				Flags: bluetooth.CharacteristicReadPermission,
				Value: []byte{0x02, 0x8a, 0x24, 0x66, 0x82, 0x34, 0x36},
				//Value: []byte{0x02, uint8(0x10C4 >> 8), uint8(0x10C4 & 0xff), uint8(0x0001 >> 8), uint8(0x0001 & 0xff)},
			},
		},
	})
	adapter.AddService(&bluetooth.Service{
		UUID: bluetooth.ServiceUUIDBattery,
		Characteristics: []bluetooth.CharacteristicConfig{
			{
				UUID:  bluetooth.CharacteristicUUIDBatteryLevel,
				Value: []byte{80},
				Flags: bluetooth.CharacteristicReadPermission | bluetooth.CharacteristicNotifyPermission,
			},
		},
	})
	// gacc
	//  device name r
	//  apperance r
	//  peripheral prefreed connection

	adapter.AddService(&bluetooth.Service{
		UUID: bluetooth.ServiceUUIDGenericAccess,
		Characteristics: []bluetooth.CharacteristicConfig{
			{
				UUID:  bluetooth.CharacteristicUUIDDeviceName,
				Flags: bluetooth.CharacteristicReadPermission,
				Value: []byte("tinygo-kinadv2"),
			},
			{

				UUID:  bluetooth.New16BitUUID(0x2A01),
				Flags: bluetooth.CharacteristicReadPermission,
				Value: []byte{uint8(0x03c4 >> 8), uint8(0x03c4 & 0xff)}, /// []byte(strconv.Itoa(961)),
			},
			// {
			// 	UUID:  bluetooth.CharacteristicUUIDPeripheralPreferredConnectionParameters,
			// 	Flags: bluetooth.CharacteristicReadPermission,
			// 	Value: []byte{0x02},
			// },

			// // 		//
		},
	})

	//v := []byte{0x85, 0x02} // 0x85, 0x02
	reportValue := []byte{0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}

	//var reportmap bluetooth.Characteristic

	// hid
	adapter.AddService(&bluetooth.Service{
		UUID: bluetooth.ServiceUUIDHumanInterfaceDevice,
		/*
			 - hid information r
			 - report map r
			 - report nr
			   - client charecteristic configuration
			   - report reference
			- report nr
			   - client charecteristic configuration
			   - report reference
			- hid control point wnr
		*/
		Characteristics: []bluetooth.CharacteristicConfig{
			// {
			// 	UUID:  bluetooth.CharacteristicUUIDHIDInformation,
			// 	Flags: bluetooth.CharacteristicReadPermission,
			// 	Value: []byte{uint8(0x0111 >> 8), uint8(0x0111 & 0xff), uint8(0x0002 >> 8), uint8(0x0002 & 0xff)},
			// },
			{
				//Handle: &reportmap,
				UUID:  bluetooth.CharacteristicUUIDReportMap,
				Flags: bluetooth.CharacteristicReadPermission,
				Value: reportMap,
			},
			{

				Handle: &ble.reportIn,
				UUID:   bluetooth.CharacteristicUUIDReport,
				Value:  reportValue[:],
				Flags:  bluetooth.CharacteristicReadPermission | bluetooth.CharacteristicNotifyPermission,
			},
			{
				// protocl mode
				UUID:  bluetooth.New16BitUUID(0x2A4E),
				Flags: bluetooth.CharacteristicWriteWithoutResponsePermission | bluetooth.CharacteristicReadPermission,
				// Value: []byte{uint8(1)},
				// WriteEvent: func(client bluetooth.Connection, offset int, value []byte) {
				// 	print("protocol mode")
				// },
			},
			{
				UUID:  bluetooth.CharacteristicUUIDHIDControlPoint,
				Flags: bluetooth.CharacteristicWriteWithoutResponsePermission,
				//	Value: []byte{0x02},
			},
		},
	})
}

func (ble *bleHost) LEDs() uint8 {
	return 0
}

func (ble *bleHost) Send(rpt keyboard.Report) {
	// if debug {
	// 	writeDebug(rpt[:])
	// }
	switch rpt.Type() {
	case keyboard.RptKeyboard:
		ble.sendKeyboardReport(rpt[0], rpt[2], rpt[3], rpt[4], rpt[5], rpt[6], rpt[7])
	case keyboard.RptMouse:
		// sendMouseReport(rpt[2], rpt[3], rpt[4], rpt[5])
		println("mouse not implemented yet")
	case keyboard.RptConsumer:
		// sendConsumerReport(uint16(rpt[3])<<8|(uint16(rpt[2])), 0, 0, 0)
		println("consumer not implemented yet")
	}
}

func (ble *bleHost) sendKeyboardReport(mod, k1, k2, k3, k4, k5, k6 byte) {
	ble.keybuf[0] = 0x01
	ble.keybuf[1] = mod
	ble.keybuf[2] = 0
	ble.keybuf[3] = k1
	ble.keybuf[4] = k2
	ble.keybuf[5] = k3
	ble.keybuf[6] = k4
	ble.keybuf[7] = k5
	ble.keybuf[8] = k6
	println("sending buffer to BLE",
		ble.keybuf[0],
		ble.keybuf[1],
		ble.keybuf[2],
		ble.keybuf[3],
		ble.keybuf[4],
		ble.keybuf[5],
		ble.keybuf[6],
		ble.keybuf[7],
		ble.keybuf[8],
	)
	if _, err := ble.reportIn.Write(ble.keybuf[:]); err != nil {
		println("failed to send key:", err.Error())
		return
	}
	println("sent keyboard report")
}

func (ble *bleHost) sendMouseReport(buttons, x, y, wheel byte) {
	// mousebuf[0] = 0x01
	// mousebuf[1] = buttons
	// mousebuf[2] = x
	// mousebuf[3] = y
	// mousebuf[4] = wheel
	// port.tx(mousebuf)
}

func (ble *bleHost) sendConsumerReport(k1, k2, k3, k4 uint16) {
	// conbuf[0] = 0x03 // REPORT_ID
	// conbuf[1] = uint8(k1)
	// conbuf[2] = uint8((k1 & 0x0300) >> 8)
	// conbuf[3] = uint8(k2)
	// conbuf[4] = uint8((k2 & 0x0300) >> 8)
	// conbuf[5] = uint8(k3)
	// conbuf[6] = uint8((k3 & 0x0300) >> 8)
	// conbuf[7] = uint8(k4)
	// conbuf[8] = uint8((k4 & 0x0300) >> 8)
	// port.tx(conbuf)
}
