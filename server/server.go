package main

import (
	"fmt"
	"log"
	"math"
	"net/http"
	"strconv"
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

func stateManager(tempUpdatesChan <-chan float64, requestsChan <-chan StateRequest, intervalUpdatesChan chan<- time.Duration) {
	const t1 float64 = 25.0
	const t2 float64 = 70.0
	const normalFreq time.Duration = 500 * time.Millisecond
	const fastFreq time.Duration = 100 * time.Millisecond
	const historySize = 100 // Manteniamo le ultime 100 letture

	state := SystemState{
		SystemStatus:     "NORMAL",
		SamplingInterval: normalFreq,
		OperativeMode:    "AUTOMATIC",
		MinTemp:          math.Inf(1),  // Inizializza min a +infinito
		MaxTemp:          math.Inf(-1), // Inizializza max a -infinito
	}
	var tempHistory []float64

	// Invia subito la frequenza iniziale al publisher
	intervalUpdatesChan <- state.SamplingInterval

	fmt.Println("INFO: State Manager avviato.")

	for {
		select {
		case req := <-requestsChan:
			req.ReplyChan <- state

		case temp := <-tempUpdatesChan:
			// Aggiorna la cronologia delle temperature
			tempHistory = append(tempHistory, temp)
			if len(tempHistory) > historySize {
				tempHistory = tempHistory[1:] // Mantiene la dimensione della cronologia
			}

			// Ricalcola le statistiche
			var sum float64
			min := math.Inf(1)
			max := math.Inf(-1)
			for _, t := range tempHistory {
				sum += t
				if t < min {
					min = t
				}
				if t > max {
					max = t
				}
			}

			// Aggiorna lo stato
			state.CurrentTemp = temp
			state.AverageTemp = sum / float64(len(tempHistory))
			state.MinTemp = min
			state.MaxTemp = max

			// Logica di stato basata sulla temperatura corrente
			var newState string
			var newFreq time.Duration

			if temp >= t2 {
				newState = "ALARM"
				newFreq = fastFreq
			} else if temp >= t1 {
				newState = "HOT-STATE"
				newFreq = fastFreq
			} else {
				newState = "NORMAL"
				newFreq = normalFreq
			}

			if newState != state.SystemStatus {
				log.Printf("ATTENZIONE: Cambio di stato -> %s (Temp: %.1f°C)", newState, temp)
				state.SystemStatus = newState
			}

			if newFreq != state.SamplingInterval {
				log.Printf("INFO: Frequenza di campionamento cambiata a %v", newFreq)
				state.SamplingInterval = newFreq
				intervalUpdatesChan <- state.SamplingInterval
			}
		}
	}
}

// startMqttPublisher ora riceve la frequenza da inviare dallo stateManager.
func startMqttPublisher(client MQTT.Client, IntervalUpdatesChan <-chan time.Duration) {
	const configTopic = "esp32/config/interval"
	fmt.Println("INFO: Publisher MQTT avviato.")

	for {
		interval := <-IntervalUpdatesChan
		intervalPayload := strconv.FormatInt(interval.Milliseconds(), 10)

		token := client.Publish(configTopic, 1, false, intervalPayload)
		token.Wait()

		if token.Error() != nil {
			log.Printf("ERRORE: Impossibile pubblicare su MQTT: %v\n", token.Error())
		} else {
			fmt.Printf("INFO: Messaggio di configurazione pubblicato su topic '%s': '%s' ms\n", configTopic, intervalPayload)
		}
	}
}

func startApiServer(requestsChan chan<- StateRequest) {
	var usemock = false
	apiController := NewController(usemock, requestsChan)

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

func main() {
	// --- Canali per la comunicazione tra goroutine ---
	tempUpdatesChan := make(chan float64)
	stateRequestsChan := make(chan StateRequest)
	intervalUpdatesChan := make(chan time.Duration)

	// --- Configurazione e connessione MQTT ---
	const broker = "tcp://localhost:1883"
	const tempTopic = "esp32/data/temperature"
	opts := MQTT.NewClientOptions().AddBroker(broker).SetClientID("go-server-logic")
	opts.SetDefaultPublishHandler(func(client MQTT.Client, msg MQTT.Message) {
		fmt.Printf("Messaggio non gestito: %s dal topic: %s\n", msg.Payload(), msg.Topic())
	})

	client := MQTT.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		log.Fatalf("ERRORE: Impossibile connettersi al broker MQTT: %v", token.Error())
	}
	fmt.Println("INFO: Connesso al broker MQTT.")

	// --- Handler per i messaggi di temperatura ---
	var temperatureMessageHandler MQTT.MessageHandler = func(client MQTT.Client, msg MQTT.Message) {
		fmt.Printf("DEBUG: Ricevuto messaggio: %s dal topic: %s\n", msg.Payload(), msg.Topic())
		temp, err := strconv.ParseFloat(string(msg.Payload()), 64)
		if err != nil {
			log.Printf("ERRORE: Impossibile convertire il payload della temperatura: %v", err)
			return
		}
		// Invia la temperatura allo stateManager
		tempUpdatesChan <- temp
	}

	// --- Sottoscrizione al topic della temperatura ---
	if token := client.Subscribe(tempTopic, 1, temperatureMessageHandler); token.Wait() && token.Error() != nil {
		log.Fatalf("ERRORE: Impossibile sottoscriversi al topic %s: %v", tempTopic, token.Error())
	}
	fmt.Printf("INFO: Sottoscritto con successo al topic: %s\n", tempTopic)

	// --- Avvio delle Goroutine ---
	go stateManager(tempUpdatesChan, stateRequestsChan, intervalUpdatesChan)
	go startApiServer(stateRequestsChan)
	go startMqttPublisher(client, intervalUpdatesChan)

	fmt.Println("INFO: Tutti i servizi sono stati avviati.")
	select {}
}
