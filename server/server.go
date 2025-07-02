package main

import (
	"fmt"
	"net/http"
	"time"
)

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

func startApiServer() {
	// Decidi quale controller usare. false per quello reale, true per il mock.
	useMockController := true
	apiController := NewController(useMockController)

	routes := map[string]http.HandlerFunc{
		"/api/temperature-stats":  apiController.TemperatureStats,
		"/api/devices-states":     apiController.DevicesStates,
		"/api/system-status":      apiController.SystemStatus,
		"/api/window-position":    apiController.WindowPosition,
		"/api/change-mode":        apiController.ChangeMode,
		"/api/open-window":        apiController.OpenWindow,
		"/api/close-window":       apiController.CloseWindow,
		"/api/reset-alarm":        apiController.ResetAlarm,
		"/api/get-alarms":         apiController.GetAlarms,
		"/api/get-operative-mode": apiController.GetOperativeMode,
	}

	for path, handler := range routes {
		http.Handle(path, corsMiddleware(http.HandlerFunc(handler)))
	}

	fmt.Println("API in ascolto su :8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Println("Errore nell'avvio del server:", err)
	}
}

func main() {
	go startApiServer()

	fmt.Println("Server avviato in background.")

	for {
		fmt.Println("Il programma principale è ancora in esecuzione...")
		time.Sleep(10 * time.Second)
	}
}
