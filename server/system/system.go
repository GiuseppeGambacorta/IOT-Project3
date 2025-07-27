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
	HotState
	Alarm
)

func (ss SystemStatus) String() string { // cosi si aggiungono a mano
	switch ss {
	case Normal:
		return "Normal"
	case HotState:
		return "HotState"
	case Alarm:
		return "Alarm"
	default:
		return ""
	}
}

// System rimane invariato.
type System struct {
	CurrentTemp      float64
	AverageTemp      float64
	MaxTemp          float64
	MinTemp          float64
	Status           SystemStatus
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
