package system

import (
	"time"
)

// SystemState rimane invariato.
type SystemState struct {
	CurrentTemp      float64
	AverageTemp      float64
	MaxTemp          float64
	MinTemp          float64
	SystemStatus     string // "NORMAL", "HOT-STATE", "ALARM"
	SamplingInterval time.Duration
	DevicesOnline    map[string]bool
	WindowPosition   int
	OperativeMode    string // "AUTOMATIC" o "MANUAL"
}

// RequestType è ancora usato per i comandi di modifica.
type RequestType int

const (
	ToggleMode RequestType = iota
	OpenWindow
	CloseWindow
	ResetAlarm
)
