package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

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
}

// --- Factory Function ---

func NewController(useMock bool) APIController {
	if useMock {
		fmt.Println("INFO: Utilizzo del controller MOCK.")
		return &MockController{}
	}
	fmt.Println("INFO: Utilizzo del controller REALE.")
	return &AppController{}
}

// --- Implementazione Reale (AppController) ---

type AppController struct {
	// Qui andranno le dipendenze reali, es. client MQTT, connessione DB, etc.
}

func (c *AppController) TemperatureStats(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Richiesta statistiche temperatura (REALE)")
	w.Header().Set("Content-Type", "application/json")
	stats := map[string]float64{"current": 22.5, "average": 21.8, "max": 25.1, "min": 19.5}
	json.NewEncoder(w).Encode(stats)
}

func (c *AppController) DevicesStates(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Richiesta stato dispositivi (REALE)")
	w.Header().Set("Content-Type", "application/json")
	states := map[string]bool{"server": true, "arduino": true, "esp32": false}
	json.NewEncoder(w).Encode(states)
}

func (c *AppController) SystemStatus(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Richiesta stato sistema (REALE)")
	w.Write([]byte("NORMAL"))
}

func (c *AppController) WindowPosition(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Richiesta posizione finestra (REALE)")
	w.Write([]byte("50"))
}

func (c *AppController) ChangeMode(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Metodo non consentito", http.StatusMethodNotAllowed)
		return
	}
	fmt.Println("Richiesta cambio modalità (REALE)")
	w.Write([]byte("OK"))
}

func (c *AppController) OpenWindow(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Metodo non consentito", http.StatusMethodNotAllowed)
		return
	}
	fmt.Println("Richiesta apertura finestra (REALE)")
	w.Write([]byte("OK"))
}

func (c *AppController) CloseWindow(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Metodo non consentito", http.StatusMethodNotAllowed)
		return
	}
	fmt.Println("Richiesta chiusura finestra (REALE)")
	w.Write([]byte("OK"))
}

func (c *AppController) ResetAlarm(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Metodo non consentito", http.StatusMethodNotAllowed)
		return
	}
	fmt.Println("Allarme resettato! (REALE)")
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
