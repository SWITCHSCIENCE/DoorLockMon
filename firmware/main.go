package main

import (
	"time"

	"DoorLockMon/firmware/core"
	"DoorLockMon/firmware/radio"
	"DoorLockMon/firmware/vdd"
	"DoorLockMon/messages"
)

func main() {
	core.Setup()
	println("wakeup")
	defer core.PowerOFF()
	ble := radio.New("Door-Lock")
	if ble.Run() {
		msg := messages.New()
		// msg.Temperature, msg.Humidity = getSensor()
		msg.Battery = vdd.Measure()
		x, y, z, err := core.Sensor().GetAccel()
		if err != nil {
			core.Failed(err)
		}
		msg.Rotation = messages.GetRotation(x, y, z)
		if err := ble.Write(msg); err != nil {
			println("write failed:", err.Error())
			return
		}
		println("report done")
	} else {
		println("authentication failed")
	}
	time.Sleep(5 * time.Second)
}
