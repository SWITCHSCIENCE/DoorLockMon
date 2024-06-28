package imu

import (
	"fmt"
	"machine"

	"tinygo.org/x/drivers/lsm303agr"
)

var (
	wire = machine.I2C0
	imu  = lsm303agr.New(wire)
)

type IMU struct {
	*lsm303agr.Device
	bus *machine.I2C
}

func New(bus *machine.I2C) (*IMU, error) {
	imu := &IMU{
		Device: lsm303agr.New(bus),
		bus:    bus,
	}
	if err := imu.Configure(lsm303agr.Configuration{
		AccelDataRate:  lsm303agr.ACCEL_DATARATE_10HZ,
		AccelPowerMode: lsm303agr.ACCEL_POWER_LOW,
		AccelRange:     lsm303agr.ACCEL_RANGE_2G,
	}); err != nil {
		return nil, err
	}
	if !imu.Connected() {
		return nil, fmt.Errorf("LSM303AGR/MAG not connected!")
	}
	if imu.ReadAccel(0x0f) != 0x33 {
		return nil, fmt.Errorf("LSM303AGR/MAG not connected!")
	}
	if imu.ReadMag(0x4f) != 0x40 {
		return nil, fmt.Errorf("LSM303AGR/MAG not connected!")
	}
	imu.WriteAccel(0x1f, 0b00000000) // TEMP_CFG_REG_A
	imu.WriteAccel(0x20, 0b00111111) // CTRL_REG1_A
	imu.WriteAccel(0x21, 0b00000000) // CTRL_REG2_A
	imu.WriteAccel(0x22, 0b01100000) // CTRL_REG3_A
	imu.WriteAccel(0x23, 0b00000000) // CTRL_REG4_A
	imu.WriteAccel(0x24, 0b00000000) // CTRL_REG5_A
	imu.WriteAccel(0x25, 0b00000000) // CTRL_REG6_A
	imu.WriteAccel(0x26, 0b00000000) // REFERENCE
	imu.WriteAccel(0x32, 0b00011100) // INT1_THS_A
	imu.WriteAccel(0x33, 0b00000100) // INT1_DURATION_A
	imu.WriteAccel(0x30, 0b01111111) // INT1_CFG_A
	imu.WriteAccel(0x34, 0b00000000) // INT2_CFG_A
	imu.WriteAccel(0x36, 0b00000000) // INT2_THS_A
	imu.WriteAccel(0x37, 0b00000000) // INT2_DURATION_A

	imu.WriteAccel(0x38, 0b00000000) // CLICK_CFG_A
	imu.WriteAccel(0x39, 0b00000000) // CLICK_SRC_A

	//imu.WriteMag(0x60, 0b00100000) // MAG_MR_REG_M
	//time.Sleep(time.Millisecond * 100)
	imu.WriteMag(0x60, 0b00000001) // MAG_MR_REG_M
	imu.WriteMag(0x61, 0b00011001) // CFG_REG_B_M
	imu.WriteMag(0x62, 0b01000000) // CFG_REG_C_M
	imu.WriteMag(0x65, 0b11111111) // INT_THS_L_REG_M
	imu.WriteMag(0x66, 0b11111111) // INT_THS_H_REG_M
	imu.WriteMag(0x63, 0b00000101) // INT_CTRL_REG_M
	imu.ReadAcceleration()
	imu.ReadMagneticField()
	imu.ReadAccel(0x31)
	imu.ReadMag(0x64)
	return imu, nil
}

func (imu *IMU) GetAccel() (float32, float32, float32, error) {
	x, y, z, err := imu.ReadAcceleration()
	if err != nil {
		return 0, 0, 0, fmt.Errorf("IMU: %w", err)
	}
	return float32(x) / 1000000, float32(y) / 1000000, float32(z) / 1000000, nil
}

func (imu *IMU) ReadAccel(reg uint8) byte {
	b := []byte{0}
	if err := imu.bus.ReadRegister(imu.AccelAddress, reg, b); err != nil {
		panic(err)
	}
	return b[0]
}

func (imu *IMU) WriteAccel(reg, data uint8) {
	if err := imu.bus.WriteRegister(imu.AccelAddress, reg, []byte{data}); err != nil {
		panic(err)
	}
}

func (imu *IMU) ReadMag(reg uint8) byte {
	b := []byte{0}
	if err := imu.bus.ReadRegister(imu.MagAddress, reg, b); err != nil {
		panic(err)
	}
	return b[0]
}

func (imu *IMU) WriteMag(reg, data uint8) {
	if err := imu.bus.WriteRegister(imu.MagAddress, reg, []byte{data}); err != nil {
		panic(err)
	}
}
