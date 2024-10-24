//go:build nrf52840

package core

import (
	"device/nrf"
	"fmt"
	"machine"
	"time"

	"DoorLockMon/firmware/imu"
)

const (
	DEBUG = false

	SW1      = machine.P0_18 // Reset
	SW2      = machine.P0_08
	LED1     = machine.P0_28 // Red
	LED2     = machine.P0_29 // Orange
	LED3     = machine.P0_30 // Green
	MOTOR    = machine.P0_31
	SDA      = machine.P0_12
	SCL      = machine.P0_14
	IMU_DRDY = machine.P0_02
	IMU_INT  = machine.P0_03
	IMU_INT1 = machine.P0_05
	IMU_DEN  = machine.P0_06
)

var (
	IMU_BUS = machine.I2C1
	sensor  *imu.IMU
	motor   uint8
)

func setMotor(d float32) {
	machine.PWM0.Set(motor, uint32(float32(machine.PWM0.PWM.GetCOUNTERTOP())*d))
}

func Setup() {
	IMU_DRDY.Configure(machine.PinConfig{Mode: machine.PinInput})
	IMU_INT.Configure(machine.PinConfig{Mode: machine.PinInput})
	IMU_INT1.Configure(machine.PinConfig{Mode: machine.PinInputPullup})
	IMU_DEN.Configure(machine.PinConfig{Mode: machine.PinInput})
	LED1.Configure(machine.PinConfig{Mode: machine.PinOutput})
	LED2.Configure(machine.PinConfig{Mode: machine.PinOutput})
	LED3.Configure(machine.PinConfig{Mode: machine.PinOutput})
	MOTOR.Configure(machine.PinConfig{Mode: machine.PinOutput})

	nrf.P0.PIN_CNF[IMU_INT1].Set(
		nrf.GPIO_PIN_CNF_DIR_Input<<nrf.GPIO_PIN_CNF_DIR_Pos |
			nrf.GPIO_PIN_CNF_PULL_Pullup<<nrf.GPIO_PIN_CNF_PULL_Pos |
			nrf.GPIO_PIN_CNF_SENSE_Low<<nrf.GPIO_PIN_CNF_SENSE_Pos,
	)
	nrf.P0.PIN_CNF[SW2].Set(
		nrf.GPIO_PIN_CNF_DIR_Input<<nrf.GPIO_PIN_CNF_DIR_Pos |
			nrf.GPIO_PIN_CNF_PULL_Pullup<<nrf.GPIO_PIN_CNF_PULL_Pos |
			nrf.GPIO_PIN_CNF_SENSE_Low<<nrf.GPIO_PIN_CNF_SENSE_Pos,
	)

	machine.PWM0.Configure(machine.PWMConfig{
		Period: 16384e3,
	})
	if err := IMU_BUS.Configure(machine.I2CConfig{
		Frequency: 400 * machine.KHz,
		SDA:       SDA,
		SCL:       SCL,
		Mode:      machine.Mode0,
	}); err != nil {
		Failed(err)
	}
	s, err := imu.New(IMU_BUS)
	if err != nil {
		Failed(err)
	}
	sensor = s
	m, err := machine.PWM0.Channel(MOTOR)
	if err != nil {
		Failed(fmt.Errorf("PWM0.Channel: %w", err))
	}
	motor = m
	LED1.Low()
	LED2.High()
	LED3.High()
	setMotor(0.0)
	if DEBUG {
		for !machine.USBCDC.DTR() {
			time.Sleep(time.Millisecond)
		}
		fmt.Println("start")
	} else {
		nrf.USBD.SetLOWPOWER(1)
		nrf.USBD.SetENABLE(0)
		nrf.UART0.TASKS_STOPTX.Set(1)
		nrf.UART0.TASKS_STOPRX.Set(1)
		nrf.UART0.SetENABLE(0)
		nrf.UARTE0.TASKS_STOPTX.Set(1)
		nrf.UARTE0.TASKS_STOPRX.Set(1)
		nrf.UARTE0.SetENABLE(0)
		nrf.UARTE1.TASKS_STOPTX.Set(1)
		nrf.UARTE1.TASKS_STOPRX.Set(1)
		nrf.UARTE1.SetENABLE(0)
	}
}

func Sensor() *imu.IMU {
	return sensor
}

func PowerOFF() {
	println("power-off")
	LED1.High()
	LED2.High()
	LED3.High()
	setMotor(0.0)
	nrf.POWER.SetSYSTEMOFF(1)
	select {}
}
