package radio

import (
	"bytes"
	"crypto/md5"
	"crypto/rand"
	"fmt"
	"time"

	"tinygo.org/x/bluetooth"

	"DoorLockMon/firmware/core"
	"DoorLockMon/messages"
	"DoorLockMon/realm"
)

var (
	adapter        = bluetooth.DefaultAdapter
	serviceUUID, _ = bluetooth.ParseUUID("a0b40001-926d-4d61-98df-8c5c62ee53b3")
	requestUUID, _ = bluetooth.ParseUUID("a0b40002-926d-4d61-98df-8c5c62ee53b4")
	reportUUID, _  = bluetooth.ParseUUID("a0b40002-926d-4d61-98df-8c5c62ee53b5")
	random         = make([]byte, 4)
	verify         = make([]byte, md5.Size)
)

func init() {
	rand.Read(random)
	if err := adapter.Enable(); err != nil {
		core.Failed(fmt.Errorf("adapter.Enable: %w", err))
	}
	hash := md5.New()
	hash.Write([]byte(realm.Value))
	hash.Write(random)
	copy(verify, hash.Sum(nil))
}

type Radio struct {
	adv       *bluetooth.Advertisement
	report    bluetooth.Characteristic
	event     chan []byte
	completed chan bool
}

// New ...
func New(name string) *Radio {
	r := &Radio{
		adv:       adapter.DefaultAdvertisement(),
		event:     make(chan []byte, 8),
		completed: make(chan bool, 1),
	}
	if err := r.adv.Configure(bluetooth.AdvertisementOptions{
		LocalName: name,
		ManufacturerData: []bluetooth.ManufacturerDataElement{
			{CompanyID: 0xffff, Data: random},
		},
	}); err != nil {
		core.Failed(fmt.Errorf("adv.Configure: %w", err))
	}
	var requestCharacteristic bluetooth.Characteristic
	if err := adapter.AddService(&bluetooth.Service{
		UUID: serviceUUID,
		Characteristics: []bluetooth.CharacteristicConfig{
			{
				Handle: &requestCharacteristic,
				UUID:   requestUUID,
				Value:  make([]byte, md5.Size),
				Flags:  bluetooth.CharacteristicWritePermission | bluetooth.CharacteristicWriteWithoutResponsePermission,
				WriteEvent: func(client bluetooth.Connection, offset int, value []byte) {
					r.event <- value
				},
			}, {
				Handle: &r.report,
				UUID:   reportUUID,
				Value:  messages.New().Bytes(),
				Flags:  bluetooth.CharacteristicReadPermission | bluetooth.CharacteristicNotifyPermission,
			},
		},
	}); err != nil {
		core.Failed(fmt.Errorf("adapter.AddService: %w", err))
	}
	return r
}

func (r *Radio) Run() bool {
	println("advertize start")
	if err := r.adv.Start(); err != nil {
		core.Failed(fmt.Errorf("adv.Start: %w", err))
	}
	timeout := time.NewTimer(30 * time.Second)
	connect := make(chan bool)
	adapter.SetConnectHandler(func(dev bluetooth.Device, connected bool) {
		println("connected")
		connect <- connected
	})
	for {
		select {
		case connected := <-connect:
			if connected {
				r.adv.Stop()
			} else {
				r.adv.Start()
			}
			timeout.Reset(25 * time.Second)
		case b := <-r.event:
			timeout.Reset(25 * time.Second)
			return bytes.Equal(b, verify)
		case <-timeout.C:
			return false
		}
	}
}

func (r *Radio) Write(msg *messages.Message) error {
	if _, err := r.report.Write(msg.Bytes()); err != nil {
		return fmt.Errorf("reportCharacteristic.Write: %w", err)
	}
	return nil
}
