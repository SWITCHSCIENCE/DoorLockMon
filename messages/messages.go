package messages

import (
	"encoding/binary"
	"strconv"
)

type Rotation uint8

const (
	RotationUnknown Rotation = iota
	RotationL
	RotationR
	RotationF
	RotationB
	RotationU
	RotationD
)

func (r Rotation) String() string {
	switch r {
	case RotationL:
		return "L"
	case RotationR:
		return "R"
	case RotationF:
		return "F"
	case RotationB:
		return "B"
	case RotationU:
		return "U"
	case RotationD:
		return "D"
	default:
		return "Unknown"
	}
}

var endian = binary.LittleEndian

type Message struct {
	Temperature float32
	Humidity    float32
	Battery     float32
	Rotation    Rotation
	buff        []byte
}

func New() *Message {
	return &Message{
		Temperature: 0,
		Humidity:    0,
		Battery:     0,
		Rotation:    RotationUnknown,
		buff:        make([]byte, 7),
	}
}

func (m *Message) Bytes() []byte {
	endian.PutUint16(m.buff[0:2], uint16(m.Temperature*100))
	endian.PutUint16(m.buff[2:4], uint16(m.Humidity*100))
	endian.PutUint16(m.buff[4:6], uint16(m.Battery*100))
	m.buff[6] = byte(m.Rotation)
	return m.buff
}

func (m *Message) Unmarshal(data []byte) error {
	m.Temperature = float32(int16(endian.Uint16(data[0:2]))) / 100
	m.Humidity = float32(int16(endian.Uint16(data[2:4]))) / 100
	m.Battery = float32(int16(endian.Uint16(data[4:6]))) / 100
	m.Rotation = Rotation(data[6])
	return nil
}

func (m *Message) String() string {
	t := strconv.FormatFloat(float64(m.Temperature), 'f', 2, 32)
	h := strconv.FormatFloat(float64(m.Humidity), 'f', 2, 32)
	b := strconv.FormatFloat(float64(m.Battery), 'f', 2, 32)
	return "T: " + t + ", H: " + h + ", B: " + b + ", R: " + m.Rotation.String()
}

func GetRotation(x, y, z float32) Rotation {
	l, r := x < -0.5, x > 0.5
	u, d := y < -0.5, y > 0.5
	f, b := z < -0.5, z > 0.5
	switch {
	case l:
		return RotationL
	case r:
		return RotationR
	case u:
		return RotationU
	case d:
		return RotationD
	case f:
		return RotationF
	case b:
		return RotationB
	}
	return RotationUnknown
}
