package led

import (
	"context"

	"tinygo.org/x/drivers/microbitmatrix"
)

type LED struct {
	device microbitmatrix.Device
	buff   []string
	cancel func()
}

func New(rotation uint8) *LED {
	device := microbitmatrix.New()
	device.Configure(microbitmatrix.Config{Rotation: rotation})
	device.ClearDisplay()
	return &LED{
		device: device,
	}
}

func (s *LED) Do(ctx context.Context) {
	defer s.device.DisableAll()
	defer s.device.ClearDisplay()

	ctx, c := context.WithCancel(ctx)
	s.cancel = c
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}
		for i := range s.buff {
			for j := range s.buff[i] {
				c := microbitmatrix.Brightness0
				if s.buff[i][j] != '0' {
					c = microbitmatrix.Brightness9
				}
				s.device.SetPixel(int16(i), int16(j), c)
			}
		}
		s.device.Display()
	}
}

func (s *LED) Stop() {
	s.cancel()
}

func (s *LED) Show(src []string) {
	s.buff = src
}
