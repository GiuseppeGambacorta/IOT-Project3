package webserver

import (
	"context"
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

func ApiServer(ctx context.Context, useMock bool, commandChan chan<- system.RequestType, stateReqChan chan<- chan system.SystemState) {
	apiController := NewController(useMock, commandChan, stateReqChan)
	routes := map[string]http.HandlerFunc{
		"/api/system-status": apiController.GetSystemStatus,
		"/api/change-mode":   apiController.ChangeMode,
		"/api/open-window":   apiController.OpenWindow,
		"/api/close-window":  apiController.CloseWindow,
		"/api/reset-alarm":   apiController.ResetAlarm,
	}
	for path, handler := range routes {
		http.Handle(path, corsMiddleware(http.HandlerFunc(handler)))
	}

	fileServer := http.FileServer(http.Dir("../dashboard-frontend"))
	http.Handle("/", fileServer)

	server := &http.Server{Addr: ":8080"}

	go func() {
		<-ctx.Done()
		log.Println("API server: Shutdown")
		server.Shutdown(context.Background()) //passo un context locale per lo shutdonw
	}()

	log.Println("INFO: API in ascolto su :8080")
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("ERRORE: Impossibile avviare il server API: %v", err)
	}
}
