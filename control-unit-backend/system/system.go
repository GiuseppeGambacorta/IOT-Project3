package system

import (
	"log"
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

type SystemStatus int

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
	CurrentTemp           float64
	AverageTemp           float64
	MaxTemp               float64
	MinTemp               float64
	Status                SystemStatus
	StatusString          string
	SamplingInterval      time.Duration
	DevicesOnline         map[DeviceName]bool
	WindowPosition        Degree
	CommandWindowPosition Degree
	OperativeMode         OperativeMode // "AUTOMATIC" o "MANUAL"
	OperativeModeString   string
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

const (
	MaxTemperatureBuffer = 100

	NoCommand      = 0
	CmdOpenWindow  = 1
	CmdCloseWindow = 2
)

func ToggleActualMode(actualSystemState *SystemState) {
	if actualSystemState.OperativeMode == Manual {
		actualSystemState.OperativeMode = Automatic
	} else {
		actualSystemState.OperativeMode = Manual
	}
	actualSystemState.OperativeModeString = actualSystemState.OperativeMode.String()
	log.Println("Modalita attuale: " + actualSystemState.OperativeModeString)
}

func ManageTemperature(temp float64, tempHistory []float64, actualSystemState *SystemState) []float64 {
	tempHistory = append(tempHistory, temp)
	if len(tempHistory) > MaxTemperatureBuffer {
		tempHistory = tempHistory[1:]
	}
	var sum float64
	for _, t := range tempHistory {
		sum += t
		if t < actualSystemState.MinTemp {
			actualSystemState.MinTemp = t
		}
		if t > actualSystemState.MaxTemp {
			actualSystemState.MaxTemp = t
		}
	}
	actualSystemState.CurrentTemp = temp
	actualSystemState.AverageTemp = sum / float64(len(tempHistory))
	return tempHistory
}

func manageMotorPosition(actualSystemState *SystemState, threshold1, threshold2 float64) {
	switch actualSystemState.Status {
	case Alarm, Too_hot:
		actualSystemState.CommandWindowPosition = 90
	case Hot:
		actualSystemState.CommandWindowPosition = Degree(
			(actualSystemState.CurrentTemp - threshold1) * (threshold2 / (threshold2 - threshold1)),
		)
	default:
		actualSystemState.CommandWindowPosition = 0
	}
}

func ManageSystemLogic(
	actualSystemState *SystemState,
	threshold1, threshold2 float64,
	normalFreq, fastFreq time.Duration,
	intervalUpdatesChan chan<- time.Duration,
	tooHotEnteredAt *time.Time,
	tooHotMaxDuration time.Duration) {

	oldStatus := actualSystemState.Status
	oldFreq := actualSystemState.SamplingInterval
	now := time.Now()

	if actualSystemState.Status != Alarm {
		if actualSystemState.CurrentTemp <= threshold1 {
			actualSystemState.Status = Normal
			actualSystemState.SamplingInterval = normalFreq
			*tooHotEnteredAt = time.Time{}
		} else if actualSystemState.CurrentTemp > threshold1 && actualSystemState.CurrentTemp <= threshold2 {
			actualSystemState.Status = Hot
			actualSystemState.SamplingInterval = fastFreq
			*tooHotEnteredAt = time.Time{}
		} else {
			actualSystemState.Status = Too_hot
			actualSystemState.SamplingInterval = fastFreq
			if tooHotEnteredAt.IsZero() {
				*tooHotEnteredAt = now
			}
			if !tooHotEnteredAt.IsZero() && now.Sub(*tooHotEnteredAt) > tooHotMaxDuration {
				actualSystemState.Status = Alarm
				log.Println("ALLARME: Temperatura troppo alta per troppo tempo! Stato -> Alarm")
				*tooHotEnteredAt = time.Time{}
			}
		}
	}

	manageMotorPosition(actualSystemState, threshold1, threshold2)
	actualSystemState.StatusString = actualSystemState.Status.String()
	actualSystemState.OperativeModeString = actualSystemState.OperativeMode.String()
	if actualSystemState.Status != oldStatus {
		log.Printf("ATTENZIONE: Cambio di stato -> %s (Temp: %.1f°C)", actualSystemState.Status.String(), actualSystemState.CurrentTemp)
	}
	if actualSystemState.SamplingInterval != oldFreq {
		log.Printf("INFO: Frequenza di campionamento cambiata a %v", actualSystemState.SamplingInterval)
		intervalUpdatesChan <- actualSystemState.SamplingInterval
	}
}
