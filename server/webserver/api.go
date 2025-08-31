package webserver

import (
	"log"
	"net/http"
	"server/system"
)

// --- Middleware ---
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

func ApiServer(useMock bool, commandChan chan<- system.RequestType, stateReqChan chan<- chan system.System) {
	apiController := NewController(useMock, commandChan, stateReqChan)
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

	fileServer := http.FileServer(http.Dir("../gui"))
	http.Handle("/", fileServer)

	log.Println("INFO: API in ascolto su :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("ERRORE: Impossibile avviare il server API: %v", err)
	}
}
