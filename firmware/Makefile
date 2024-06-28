TARGET=microbit-v2-s113v7
PORT=/dev/tty.usbmodem14102
PORT=COM22

test:
	echo $(SD)

build:
	tinygo build -target $(TARGET) -o main.hex .

flash:
	tinygo flash -target $(TARGET) -programmer cmsis-dap -x .

softdevice:
	cd "$(shell go list -f '{{.Dir}}' tinygo.org/x/bluetooth)/s113_nrf52_7.0.1" && \
	openocd -f interface/cmsis-dap.cfg -c 'transport select swd' -f target/nrf52.cfg -c 'program ./s113_nrf52_7.0.1_softdevice.hex reset exit'

mon:
	tinygo monitor -port $(PORT)
