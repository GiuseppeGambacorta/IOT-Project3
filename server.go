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
	// TODO: Implementare logica reale
	stats := map[string]float64{"current": 22.5, "average": 21.8, "max": 25.1, "min": 19.5}
	json.NewEncoder(w).Encode(stats)
}

func (c *AppController) DevicesStates(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Richiesta stato dispositivi (REALE)")
	w.Header().Set("Content-Type", "application/json")
	// TODO: Implementare logica reale
	states := map[string]bool{"server": true, "arduino": true, "esp32": false}
	json.NewEncoder(w).Encode(states)
}

func (c *AppController) SystemStatus(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Richiesta stato sistema (REALE)")
	w.Write([]byte("NORMAL")) // TODO: Implementare logica reale
}

func (c *AppController) WindowPosition(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Richiesta posizione finestra (REALE)")
	w.Write([]byte("50")) // TODO: Implementare logica reale
}

func (c *AppController) ChangeMode(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Metodo non consentito", http.StatusMethodNotAllowed)
		return
	}
	fmt.Println("Richiesta cambio modalità (REALE)")
	w.Write([]byte("OK")) // TODO: Implementare logica reale
}

func (c *AppController) OpenWindow(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Metodo non consentito", http.StatusMethodNotAllowed)
		return
	}
	fmt.Println("Richiesta apertura finestra (REALE)")
	w.Write([]byte("OK")) // TODO: Implementare logica reale
}

func (c *AppController) CloseWindow(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Metodo non consentito", http.StatusMethodNotAllowed)
		return
	}
	fmt.Println("Richiesta chiusura finestra (REALE)")
	w.Write([]byte("OK")) // TODO: Implementare logica reale
}

func (c *AppController) ResetAlarm(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Metodo non consentito", http.StatusMethodNotAllowed)
		return
	}
	fmt.Println("Allarme resettato! (REALE)")
	w.Write([]byte("OK")) // TODO: Implementare logica reale
}

// --- Implementazione Mock (MockController) ---

type MockController struct{}

func (c *MockController) TemperatureStats(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Richiesta statistiche temperatura (MOCK)")
	w.Header().Set("Content-Type", "application/json")
	stats := map[string]float64{"current": 99.9, "average": 99.9, "max": 99.9, "min": 99.9}
	json.NewEncoder(w).Encode(stats)
}
func (c *MockController) DevicesStates(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Richiesta stato dispositivi (MOCK)")
	w.Header().Set("Content-Type", "application/json")
	states := map[string]bool{"server": true, "arduino": true, "esp32": true}
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

// --- Middleware e Main ---

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func main() {
	// Decidi quale controller usare. false per quello reale, true per il mock.
	useMockController := false
	apiController := NewController(useMockController)

	routes := map[string]http.HandlerFunc{
		"/api/temperature-stats": apiController.TemperatureStats,
		"/api/devices-states":    apiController.DevicesStates,
		"/api/system-status":     apiController.SystemStatus,
		"/api/window-position":   apiController.WindowPosition,
		"/api/change-mode":       apiController.ChangeMode,
		"/api/open-window":       apiController.OpenWindow,
		"/api/close-window":      apiController.CloseWindow,
		"/api/reset-alarm":       apiController.ResetAlarm,
	}

	for path, handler := range routes {
		http.Handle(path, corsMiddleware(http.HandlerFunc(handler)))
	}

	fmt.Println("API in ascolto su :8080")
	http.ListenAndServe(":8080", nil)
}
