package main

import (
	"context"
	"device/nrf"
	"fmt"
	"machine"
	"time"

	"tinygo.org/x/drivers/microbitmatrix"

	"DoorLockMon/firmware/core"
	"DoorLockMon/firmware/imu"
	"DoorLockMon/firmware/led"
	"DoorLockMon/firmware/radio"
	"DoorLockMon/firmware/vdd"
	"DoorLockMon/messages"
)

const (
	MIC_IN      = machine.Pin(5)
	LOGO        = machine.Pin(36)
	SPEAKER     = machine.Pin(0)
	I2C_INT_INT = machine.Pin(25)
)

var (
	wire    = machine.I2C0
	sensor  *imu.IMU
	display = led.New(microbitmatrix.Rotation270)
)

func init() {
	ctx := context.Background()
	go display.Do(ctx)
	if err := wire.Configure(machine.I2CConfig{}); err != nil {
		core.Failed(err)
	}
	time.Sleep(time.Millisecond * 5)
	s, err := imu.New(wire)
	if err != nil {
		core.Failed(err)
	}
	sensor = s
}

func PowerOFF() {
	println("power-off")
	display.Stop()
	nrf.POWER.SetSYSTEMOFF(1)
	select {}
}

func update() messages.Rotation {
	b := sensor.ReadAccel(0x31)
	x, y, z, err := sensor.GetAccel()
	if err != nil {
		core.Failed(err)
	}
	rotation := messages.GetRotation(x, y, z)
	fmt.Printf("Interrupt: %.1f,%.1f,%.1f,0x%x,%d\n", x, y, z, b, rotation)
	if rotation == messages.RotationF {
		display.Show([]string{
			"01110",
			"10001",
			"10001",
			"10001",
			"01110",
		})
	} else {
		display.Show([]string{
			"10001",
			"01010",
			"00100",
			"01010",
			"10001",
		})
	}
	return rotation
}

func main() {
	if err := wire.WriteRegister(0x70, 0x00, nil); err != nil {
		println(err.Error())
	}
	time.Sleep(100 * time.Millisecond)
	b := make([]byte, 10)
	if err := wire.WriteRegister(0x70, 0x10, []byte{0x01}); err != nil {
		println(err.Error())
	}
	time.Sleep(20 * time.Millisecond)
	if err := wire.ReadRegister(0x70, 0x11, b[:5]); err != nil {
		println(err.Error())
	}
	fmt.Printf("%X\n", b[:5])
	if err := wire.WriteRegister(0x70, 0x12, []byte{0x07, 0x01, 0x08}); err != nil {
		println(err.Error())
	}
	time.Sleep(20 * time.Millisecond)
	if err := wire.ReadRegister(0x70, 0x13, b[:2]); err != nil {
		println(err.Error())
	}
	fmt.Printf("%X\n", b[:2])
	if err := wire.WriteRegister(0x70, 0x12, []byte{0x08, 0x01, 0x00}); err != nil {
		println(err.Error())
	}
	time.Sleep(20 * time.Millisecond)
	if err := wire.ReadRegister(0x70, 0x13, b[:2]); err != nil {
		println(err.Error())
	}
	fmt.Printf("%X\n", b[:2])
	println("wakeup", sensor.ReadAccel(0x31))
	update()
	nrf.P0.PIN_CNF[I2C_INT_INT].Set(
		nrf.GPIO_PIN_CNF_DIR_Input<<nrf.GPIO_PIN_CNF_DIR_Pos |
			nrf.GPIO_PIN_CNF_PULL_Pullup<<nrf.GPIO_PIN_CNF_PULL_Pos |
			nrf.GPIO_PIN_CNF_SENSE_Low<<nrf.GPIO_PIN_CNF_SENSE_Pos,
	)
	ch := make(chan bool, 1)
	I2C_INT_INT.SetInterrupt(machine.PinFalling, func(pin machine.Pin) {
		ch <- pin.Get()
	})
	go func() {
		for {
			<-ch
			update()
		}
	}()
	nrf.P0.PIN_CNF[machine.BUTTONA].Set(
		nrf.GPIO_PIN_CNF_DIR_Input<<nrf.GPIO_PIN_CNF_DIR_Pos |
			nrf.GPIO_PIN_CNF_PULL_Pullup<<nrf.GPIO_PIN_CNF_PULL_Pos |
			nrf.GPIO_PIN_CNF_SENSE_Low<<nrf.GPIO_PIN_CNF_SENSE_Pos,
	)
	defer PowerOFF()
	ble := radio.New("Door-Lock2")
	if ble.Run() {
		msg := messages.New()
		// msg.Temperature, msg.Humidity = getSensor()
		msg.Battery = vdd.Measure()
		x, y, z, err := sensor.GetAccel()
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
