package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// --- Interfaccia e Implementazione del Controller ---

// APIController definisce l'interfaccia per i nostri gestori di richieste.
// In questo modo possiamo facilmente creare un mock per i test.
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

// MockedController è l'implementazione concreta di APIController.
type MockedController struct {
	// In futuro, qui potresti aggiungere dipendenze come
	// un client per il database, un client MQTT, ecc.
}

// NewController crea una nuova istanza del nostro controller.
func NewController() *MockedController {
	return &MockedController{}
}

// --- Metodi del Controller (Implementazione dell'interfaccia) ---

func (c *MockedController) TemperatureStats(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Richiesta statistiche temperatura")
	w.Header().Set("Content-Type", "application/json")
	stats := map[string]float64{"current": 22.5, "average": 21.8, "max": 25.1, "min": 19.5}
	json.NewEncoder(w).Encode(stats)
}

func (c *MockedController) DevicesStates(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Richiesta stato dispositivi")
	w.Header().Set("Content-Type", "application/json")
	states := map[string]bool{"server": true, "arduino": true, "esp32": false}
	json.NewEncoder(w).Encode(states)
}

func (c *MockedController) SystemStatus(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Richiesta stato sistema")
	w.Write([]byte("NORMAL"))
}

func (c *MockedController) WindowPosition(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Richiesta posizione finestra")
	w.Write([]byte("50"))
}

func (c *MockedController) ChangeMode(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		fmt.Println("Richiesta cambio modalità")
		w.Write([]byte("OK"))
	} else {
		http.Error(w, "Metodo non consentito", http.StatusMethodNotAllowed)
	}
}

func (c *MockedController) OpenWindow(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		fmt.Println("Richiesta apertura finestra")
		w.Write([]byte("OK"))
	} else {
		http.Error(w, "Metodo non consentito", http.StatusMethodNotAllowed)
	}
}

func (c *MockedController) CloseWindow(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		fmt.Println("Richiesta chiusura finestra")
		w.Write([]byte("OK"))
	} else {
		http.Error(w, "Metodo non consentito", http.StatusMethodNotAllowed)
	}
}

func (c *MockedController) ResetAlarm(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		fmt.Println("Allarme resettato!")
		w.Write([]byte("OK"))
	} else {
		http.Error(w, "Metodo non consentito", http.StatusMethodNotAllowed)
	}
}

// --- Middleware e Main ---

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, hx-request")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func main() {
	// Crea un'istanza del controller
	apiController := NewController()

	// Definisci tutti gli endpoint e collegali ai metodi del controller
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

	// Registra tutti gli endpoint iterando sulla mappa
	for path, handler := range routes {
		http.Handle(path, corsMiddleware(http.HandlerFunc(handler)))
	}

	fmt.Println("API in ascolto su :8080")
	http.ListenAndServe(":8080", nil)
}
