package webserver

import (
	"encoding/json"
	"fmt"
	"net/http"
	"server/system"
)

type APIController interface {
	TemperatureStats(w http.ResponseWriter, r *http.Request)
	DevicesStates(w http.ResponseWriter, r *http.Request)
	SystemStatus(w http.ResponseWriter, r *http.Request)
	WindowPosition(w http.ResponseWriter, r *http.Request)
	ChangeMode(w http.ResponseWriter, r *http.Request)
	OpenWindow(w http.ResponseWriter, r *http.Request)
	CloseWindow(w http.ResponseWriter, r *http.Request)
	ResetAlarm(w http.ResponseWriter, r *http.Request)
	GetAlarms(w http.ResponseWriter, r *http.Request)
	GetOperativeMode(w http.ResponseWriter, r *http.Request)
}

func NewController(useMock bool, commandChan chan<- system.RequestType, stateReqChan chan<- chan system.System) APIController {
	if useMock {
		fmt.Println("INFO: Utilizzo del controller MOCK.")
		return &MockController{}
	}
	fmt.Println("INFO: Utilizzo del controller REALE.")
	return &AppController{
		commandChan:  commandChan,
		stateReqChan: stateReqChan,
	}
}

// --- Implementazione Reale

type AppController struct {
	commandChan  chan<- system.RequestType
	stateReqChan chan<- chan system.System
}

// getState è una funzione helper per ridurre la duplicazione di codice nelle richieste di lettura.
func (c *AppController) getState() system.System {
	replyChan := make(chan system.System)
	c.stateReqChan <- replyChan
	return <-replyChan
}

func (c *AppController) TemperatureStats(w http.ResponseWriter, r *http.Request) {
	state := c.getState()
	stats := map[string]float64{
		"current": state.CurrentTemp,
		"average": state.AverageTemp,
		"max":     state.MaxTemp,
		"min":     state.MinTemp,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

func (c *AppController) DevicesStates(w http.ResponseWriter, r *http.Request) {
	state := c.getState()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(state.DevicesOnline)
}

func (c *AppController) SystemStatus(w http.ResponseWriter, r *http.Request) {
	state := c.getState()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": state.Status.String()})
}

func (c *AppController) WindowPosition(w http.ResponseWriter, r *http.Request) {
	state := c.getState()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]system.Degree{"position": state.WindowPosition})
}

func (c *AppController) GetAlarms(w http.ResponseWriter, r *http.Request) {
	systemState := c.getState()
	isAlarmActive := systemState.Status == system.Alarm
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"attivo": isAlarmActive})
}

func (c *AppController) GetOperativeMode(w http.ResponseWriter, r *http.Request) {
	state := c.getState()
	isManual := state.OperativeMode == system.Manual
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"manuale": isManual})
}

// --- Metodi di scrittura (modificati per usare commandChan) ---
func (c *AppController) ChangeMode(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Metodo non consentito", http.StatusMethodNotAllowed)
		return
	}
	c.commandChan <- system.ToggleMode
	fmt.Println("INFO: Inviato comando di cambio modalità.")
	w.WriteHeader(http.StatusOK)
}

func (c *AppController) OpenWindow(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Metodo non consentito", http.StatusMethodNotAllowed)
		return
	}
	c.commandChan <- system.OpenWindow
	fmt.Println("INFO: Inviato comando di apertura finestra.")
	w.WriteHeader(http.StatusOK)
}

func (c *AppController) CloseWindow(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Metodo non consentito", http.StatusMethodNotAllowed)
		return
	}
	c.commandChan <- system.CloseWindow
	fmt.Println("INFO: Inviato comando di chiusura finestra.")
	w.WriteHeader(http.StatusOK)
}

func (c *AppController) ResetAlarm(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Metodo non consentito", http.StatusMethodNotAllowed)
		return
	}
	c.commandChan <- system.ResetAlarm
	fmt.Println("INFO: Inviato comando di reset allarme.")
	w.WriteHeader(http.StatusOK)
}

// --- Implementazione Mock  ---

type MockController struct{}

func (c *MockController) TemperatureStats(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Richiesta statistiche temperatura (MOCK)")
	w.Header().Set("Content-Type", "application/json")
	stats := map[string]float64{"current": 25.9, "average": 17.9, "max": 50.9, "min": 0.9}
	json.NewEncoder(w).Encode(stats)
}

func (c *MockController) DevicesStates(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Richiesta stato dispositivi (MOCK)")
	w.Header().Set("Content-Type", "application/json")
	states := map[string]bool{"server": true, "arduino": false, "esp32": true}
	json.NewEncoder(w).Encode(states)
}

func (c *MockController) SystemStatus(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Richiesta stato sistema (MOCK)")
	w.Write([]byte("MOCK_STATUS"))
}

func (c *MockController) WindowPosition(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Richiesta posizione finestra (MOCK)")
	w.Write([]byte("100"))
}

func (c *MockController) ChangeMode(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Richiesta cambio modalità (MOCK)")
	w.Write([]byte("OK"))
}

func (c *MockController) OpenWindow(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Richiesta apertura finestra (MOCK)")
	w.Write([]byte("OK"))
}

func (c *MockController) CloseWindow(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Richiesta chiusura finestra (MOCK)")
	w.Write([]byte("OK"))
}

func (c *MockController) ResetAlarm(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Allarme resettato! (MOCK)")
	w.Write([]byte("OK"))
}

func (c *MockController) GetAlarms(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Richiesta stato allarmi (MOCK)")
	w.Header().Set("Content-Type", "application/json")
	states := map[string]bool{"attivo": true}
	json.NewEncoder(w).Encode(states)
}

func (c *MockController) GetOperativeMode(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Richiesta stato impianto (MOCK)")
	w.Header().Set("Content-Type", "application/json")
	states := map[string]bool{"manuale": true}
	json.NewEncoder(w).Encode(states)
}
