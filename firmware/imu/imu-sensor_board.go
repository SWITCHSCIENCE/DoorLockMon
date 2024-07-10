//go:build nrf52840

package imu

import (
	"fmt"
	"machine"

	"tinygo.org/x/drivers/lsm9ds1"
)

const (
	SDA      = machine.P0_12
	SCL      = machine.P0_14
	IMU_DRDY = machine.P0_02
	IMU_INT  = machine.P0_03
	IMU_INT1 = machine.P0_05
	IMU_DEN  = machine.P0_06
)

type IMU struct {
	*lsm9ds1.Device
	bus        *machine.I2C
	dx, dy, dz float32
}

func New(bus *machine.I2C) (*IMU, error) {
	imu := &IMU{
		Device: lsm9ds1.New(bus),
		bus:    bus,
		dx:     1.0,
		dy:     1.0,
		dz:     -1.0,
	}
	imu.AccelAddress = 0x6a
	imu.MagAddress = 0x1c
	if err := imu.Configure(lsm9ds1.Configuration{
		AccelRange:      lsm9ds1.ACCEL_2G,
		AccelSampleRate: lsm9ds1.ACCEL_SR_10,
		GyroRange:       lsm9ds1.GYRO_250DPS,
		GyroSampleRate:  lsm9ds1.GYRO_SR_OFF,
		MagRange:        lsm9ds1.MAG_4G,
		MagSampleRate:   lsm9ds1.MAG_SR_06,
	}); err != nil {
		return nil, fmt.Errorf("IMU Setup: %w", err)
	}
	// Gyro Power OFF
	if err := imu.bus.WriteRegister(imu.MagAddress, lsm9ds1.CTRL_REG3_M, []byte{0b00000011}); err != nil {
		return nil, fmt.Errorf("IMU Setup: %w", err)
	}
	fmt.Println("imu connect:", imu.Connected())
	imu.WriteAccel(0x22, 0b00110100)
	imu.WriteAccel(0x06, 0b01111111) // Set 6D Interrupt
	imu.WriteAccel(0x07, 0b01000001)
	imu.WriteAccel(0x08, 0b01000001)
	imu.WriteAccel(0x09, 0b01000001)
	imu.WriteAccel(0x0a, 0b10000100) // INT_GEN_DUR_XL
	imu.WriteAccel(0x0c, 0b01000000)
	imu.WriteAccel(0x26, 0b01111111)
	return imu, nil
}
