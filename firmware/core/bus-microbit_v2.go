//go:build microbit_v2

package core

import (
	"context"
	"device/nrf"
	"fmt"
	"machine"
	"time"

	"tinygo.org/x/drivers/microbitmatrix"

	"DoorLockMon/firmware/imu"
	"DoorLockMon/firmware/led"
	"DoorLockMon/messages"
)

const (
	MIC_IN      = machine.P0_05
	LOGO        = machine.P1_04
	SPEAKER     = machine.P0_00
	I2C_INT_INT = machine.P0_25
)

var (
	IMU_BUS = machine.I2C0
	wire    = machine.I2C0
	display = led.New(microbitmatrix.Rotation270)
	sensor  *imu.IMU
)

func Setup() {
	if err := IMU_BUS.Configure(machine.I2CConfig{}); err != nil {
		Failed(err)
	}
	s, err := imu.New(IMU_BUS)
	if err != nil {
		Failed(err)
	}
	sensor = s
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
	ctx := context.Background()
	go display.Do(ctx)
	nrf.P0.PIN_CNF[I2C_INT_INT].Set(
		nrf.GPIO_PIN_CNF_DIR_Input<<nrf.GPIO_PIN_CNF_DIR_Pos |
			nrf.GPIO_PIN_CNF_PULL_Pullup<<nrf.GPIO_PIN_CNF_PULL_Pos |
			nrf.GPIO_PIN_CNF_SENSE_Low<<nrf.GPIO_PIN_CNF_SENSE_Pos,
	)
	nrf.P0.PIN_CNF[machine.BUTTONA].Set(
		nrf.GPIO_PIN_CNF_DIR_Input<<nrf.GPIO_PIN_CNF_DIR_Pos |
			nrf.GPIO_PIN_CNF_PULL_Pullup<<nrf.GPIO_PIN_CNF_PULL_Pos |
			nrf.GPIO_PIN_CNF_SENSE_Low<<nrf.GPIO_PIN_CNF_SENSE_Pos,
	)
}

func update() messages.Rotation {
	b := sensor.ReadAccel(0x31)
	x, y, z, err := sensor.GetAccel()
	if err != nil {
		Failed(err)
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

func Sensor() *imu.IMU {
	return sensor
}

func PowerOFF() {
	println("power-off")
	display.Stop()
	nrf.POWER.SetSYSTEMOFF(1)
	select {}
}
