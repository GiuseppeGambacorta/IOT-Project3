package system

import (
	"time"
)

type DeviceName string
type Degree int

//
//go:generate stringer -type=OperativeMode
type OperativeMode int

const (
	Manual OperativeMode = iota
	Automatic
)

type SystemStatus int16

const (
	Normal SystemStatus = iota
	Hot
	Too_hot
	Alarm
)

func (ss SystemStatus) String() string { // cosi si aggiungono a mano
	switch ss {
	case Normal:
		return "NORMAL"
	case Hot:
		return "HOT"
	case Too_hot:
		return "TOO-HOT"
	case Alarm:
		return "ALARM"
	default:
		return ""
	}
}

// System rimane invariato.
type SystemState struct {
	CurrentTemp      float64
	AverageTemp      float64
	MaxTemp          float64
	MinTemp          float64
	Status           SystemStatus
	StatusString     string
	SamplingInterval time.Duration
	DevicesOnline    map[DeviceName]bool
	WindowPosition   Degree
	OperativeMode    OperativeMode // "AUTOMATIC" o "MANUAL"
}

//
//go:generate stringer -type=RequestType
type RequestType int

const (
	ToggleMode RequestType = iota
	OpenWindow
	CloseWindow
	ResetAlarm
)
