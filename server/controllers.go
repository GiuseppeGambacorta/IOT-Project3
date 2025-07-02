package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// SystemState contiene tutti i dati condivisi dell'applicazione.
// Ho aggiunto OperativeMode per gestire la modalità manuale/automatica.
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

type StateRequest struct {
	ReplyChan chan SystemState
}

// --- Interfaccia Controller ---

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

// --- Factory Function ---

func NewController(useMock bool, requests chan<- StateRequest) APIController {
	if useMock {
		fmt.Println("INFO: Utilizzo del controller MOCK.")
		return &MockController{}
	}
	fmt.Println("INFO: Utilizzo del controller REALE.")
	return &AppController{requestsChan: requests}
}

// --- Implementazione Reale (AppController) ---

type AppController struct {
	requestsChan chan<- StateRequest
}

func (c *AppController) TemperatureStats(w http.ResponseWriter, r *http.Request) {
	replyChan := make(chan SystemState)
	c.requestsChan <- StateRequest{ReplyChan: replyChan}
	state := <-replyChan

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
	replyChan := make(chan SystemState)
	c.requestsChan <- StateRequest{ReplyChan: replyChan}
	state := <-replyChan

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(state.DevicesOnline)
}

func (c *AppController) SystemStatus(w http.ResponseWriter, r *http.Request) {
	replyChan := make(chan SystemState)
	c.requestsChan <- StateRequest{ReplyChan: replyChan}
	state := <-replyChan

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": state.SystemStatus})
}

func (c *AppController) WindowPosition(w http.ResponseWriter, r *http.Request) {
	replyChan := make(chan SystemState)
	c.requestsChan <- StateRequest{ReplyChan: replyChan}
	state := <-replyChan

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]int{"position": state.WindowPosition})
}

func (c *AppController) GetAlarms(w http.ResponseWriter, r *http.Request) {
	replyChan := make(chan SystemState)
	c.requestsChan <- StateRequest{ReplyChan: replyChan}
	state := <-replyChan

	isAlarmActive := state.SystemStatus == "ALARM"
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"attivo": isAlarmActive})
}

func (c *AppController) GetOperativeMode(w http.ResponseWriter, r *http.Request) {
	replyChan := make(chan SystemState)
	c.requestsChan <- StateRequest{ReplyChan: replyChan}
	state := <-replyChan

	isManual := state.OperativeMode == "MANUAL"
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"manuale": isManual})
}

// --- Metodi di scrittura (non ancora implementati con logica reale) ---

func (c *AppController) ChangeMode(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Metodo non consentito", http.StatusMethodNotAllowed)
		return
	}
	fmt.Println("Richiesta cambio modalità (REALE) - Logica da implementare")
	w.Write([]byte("OK"))
}

func (c *AppController) OpenWindow(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Metodo non consentito", http.StatusMethodNotAllowed)
		return
	}
	fmt.Println("Richiesta apertura finestra (REALE) - Logica da implementare")
	w.Write([]byte("OK"))
}

func (c *AppController) CloseWindow(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Metodo non consentito", http.StatusMethodNotAllowed)
		return
	}
	fmt.Println("Richiesta chiusura finestra (REALE) - Logica da implementare")
	w.Write([]byte("OK"))
}

func (c *AppController) ResetAlarm(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Metodo non consentito", http.StatusMethodNotAllowed)
		return
	}
	fmt.Println("Allarme resettato! (REALE) - Logica da implementare")
	w.Write([]byte("OK"))
}

// --- Implementazione Mock (MockController) ---

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
