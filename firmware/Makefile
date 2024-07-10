ifeq ($(ORIG),"")
TARGET=microbit-v2-s113v7
PROG=-programmer cmsis-dap
else
TARGET=./config.json
PROG=
endif
ifeq ($(OS),Windows_NT)
PORT=COM22
else
PORT=/dev/tty.usbmodem14102
endif

test:
	echo $(SD)

build:
	tinygo build -target $(TARGET) -o main.hex .

flash:
	tinygo flash -target $(TARGET) $(PROG) -x .

softdevice:
	cd "$(shell go list -f '{{.Dir}}' tinygo.org/x/bluetooth)/s113_nrf52_7.0.1" && \
	openocd -f interface/cmsis-dap.cfg -c 'transport select swd' -f target/nrf52.cfg -c 'program ./s113_nrf52_7.0.1_softdevice.hex reset exit'

mon:
	tinygo monitor -port $(PORT)
