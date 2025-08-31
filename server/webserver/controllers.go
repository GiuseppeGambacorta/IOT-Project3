package webserver

import (
	"encoding/json"
	"fmt"
	"net/http"
	"server/system"
)

type APIController interface {
	GetSystemStatus(w http.ResponseWriter, r *http.Request)
	ChangeMode(w http.ResponseWriter, r *http.Request)
	OpenWindow(w http.ResponseWriter, r *http.Request)
	CloseWindow(w http.ResponseWriter, r *http.Request)
	ResetAlarm(w http.ResponseWriter, r *http.Request)
}

func NewController(useMock bool, commandChan chan<- system.RequestType, stateReqChan chan<- chan system.SystemState) APIController {
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
	stateReqChan chan<- chan system.SystemState
}

// getState è una funzione helper per ridurre la duplicazione di codice nelle richieste di lettura.
// crea un canale e lo invia tramite il canale di comunicazione, poi aspetto la risposta sul canale inviato,
func (c *AppController) getState() system.SystemState {
	replyChan := make(chan system.SystemState)
	c.stateReqChan <- replyChan
	return <-replyChan
}

func (c *AppController) GetSystemStatus(w http.ResponseWriter, r *http.Request) {
	actualSystemState := c.getState()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(actualSystemState)
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

func (c *MockController) GetSystemStatus(w http.ResponseWriter, r *http.Request) {
	actualSystemState := system.SystemState{
		Status:           system.Normal,
		StatusString:     system.Normal.String(),
		SamplingInterval: 100,
		OperativeMode:    system.Automatic,
		CurrentTemp:      25,
		AverageTemp:      32,
		MinTemp:          47,
		MaxTemp:          12,
		DevicesOnline: map[system.DeviceName]bool{
			"server":  true,
			"esp32":   false,
			"arduino": false,
		},
	}

	fmt.Println("Richiesta stato sistema (MOCK)")
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(actualSystemState)
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
