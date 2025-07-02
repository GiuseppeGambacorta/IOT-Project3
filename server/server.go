package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	MQTT "github.com/eclipse/paho.mqtt.golang"
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

func startApiServer() {
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

	fmt.Println("INFO: API in ascolto su :8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatalf("ERRORE: Impossibile avviare il server API: %v", err)
	}
}

func startMqttPublisher(client MQTT.Client) {
	const configTopic = "esp32/config/interval"
	const intervalPayload = "100"

	fmt.Println("INFO: Publisher MQTT avviato.")

	for {
		token := client.Publish(configTopic, 1, false, intervalPayload)
		token.Wait()

		if token.Error() != nil {
			log.Printf("ERRORE: Impossibile pubblicare su MQTT: %v\n", token.Error())
		} else {
			fmt.Printf("INFO: Messaggio pubblicato su topic '%s': '%s'\n", configTopic, intervalPayload)
		}

		time.Sleep(30 * time.Second)
	}
}

func main() {
	const broker = "tcp://localhost:1883"
	opts := MQTT.NewClientOptions().AddBroker(broker)
	opts.SetClientID("go-server-publisher")
	opts.SetDefaultPublishHandler(func(client MQTT.Client, msg MQTT.Message) {
		fmt.Printf("Messaggio non gestito: %s dal topic: %s\n", msg.Payload(), msg.Topic())
	})

	client := MQTT.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		log.Fatalf("ERRORE: Impossibile connettersi al broker MQTT: %v", token.Error())
	}
	fmt.Println("INFO: Connesso al broker MQTT.")

	go startApiServer()
	go startMqttPublisher(client)

	fmt.Println("INFO: Server API e Publisher MQTT avviati in background.")

	select {}
}
