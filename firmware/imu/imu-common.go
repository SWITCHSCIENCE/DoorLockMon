package imu

import (
	"fmt"
)

func (imu *IMU) GetAccel() (float32, float32, float32, error) {
	x, y, z, err := imu.ReadAcceleration()
	if err != nil {
		return 0, 0, 0, fmt.Errorf("IMU: %w", err)
	}
	return imu.dx * float32(x) / 1000000, imu.dy * float32(y) / 1000000, imu.dz * float32(z) / 1000000, nil
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
