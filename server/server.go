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

// stateManager è la goroutine centrale che gestisce lo stato dell'applicazione.
// Ora riceve due canali separati per comandi e richieste di stato.
func stateManager(
	tempUpdatesChan <-chan float64,
	commandChan <-chan RequestType,
	stateReqChan <-chan chan SystemState,
	intervalUpdatesChan chan<- time.Duration,
) {
	const t1 float64 = 25.0
	const t2 float64 = 70.0
	const normalFreq time.Duration = 500 * time.Millisecond
	const fastFreq time.Duration = 100 * time.Millisecond
	const historySize = 100
	const esp32Timeout = 2 * time.Second

	var resetAlarm = false
	var inAlarm = false

	var newState string
	var newFreq time.Duration

	state := SystemState{
		SystemStatus:     "NORMAL",
		SamplingInterval: normalFreq,
		OperativeMode:    "AUTOMATIC",
		MinTemp:          math.Inf(1),
		MaxTemp:          math.Inf(-1),
		DevicesOnline: map[string]bool{
			"server":  true,
			"esp32":   false,
			"arduino": false,
		},
	}
	var tempHistory []float64

	esp32Timer := time.NewTimer(esp32Timeout)
	if !esp32Timer.Stop() {
		<-esp32Timer.C
	}

	intervalUpdatesChan <- state.SamplingInterval
	fmt.Println("INFO: State Manager avviato.")

	for {
		select {
		// Case per i comandi di scrittura (non richiedono risposta).
		case cmd := <-commandChan:
			switch cmd {
			case RequestToggleMode:
				if state.OperativeMode == "AUTOMATIC" {
					state.OperativeMode = "MANUAL"
				} else {
					state.OperativeMode = "AUTOMATIC"
				}
				log.Printf("INFO: Modalità operativa cambiata a %s.", state.OperativeMode)
			case RequestOpenWindow:
				log.Println("COMANDO: Apertura finestra (logica da implementare).")
			case RequestCloseWindow:
				log.Println("COMANDO: Chiusura finestra (logica da implementare).")
			case RequestResetAlarm:
				resetAlarm = true
			}

		// Case per le richieste di lettura dello stato (richiedono risposta).
		case replyChan := <-stateReqChan:
			replyChan <- state

		// Case per gli aggiornamenti di temperatura da MQTT.
		case temp := <-tempUpdatesChan:
			if !state.DevicesOnline["esp32"] {
				log.Println("INFO: Dispositivo ESP32 è ora ONLINE.")
				state.DevicesOnline["esp32"] = true
			}
			// Resetta sempre il timer quando arriva un messaggio.
			if !esp32Timer.Stop() {
				// Assicura che il canale del timer sia vuoto se era già scaduto.
				select {
				case <-esp32Timer.C:
				default:
				}
			}
			esp32Timer.Reset(esp32Timeout)

			// Aggiorna la cronologia e le statistiche delle temperature.
			tempHistory = append(tempHistory, temp)
			if len(tempHistory) > historySize {
				tempHistory = tempHistory[1:]
			}
			var sum float64
			min, max := math.Inf(1), math.Inf(-1)
			for _, t := range tempHistory {
				sum += t
				if t < min {
					min = t
				}
				if t > max {
					max = t
				}
			}
			state.CurrentTemp = temp
			state.AverageTemp = sum / float64(len(tempHistory))
			state.MinTemp = min
			state.MaxTemp = max

			// Logica di gestione dell'allarme.
			if resetAlarm {
				resetAlarm = false
				if temp <= t2 {
					inAlarm = false
				}
			}

			if !inAlarm {
				if temp <= t1 {
					newState = "NORMAL"
					newFreq = normalFreq
				} else if temp > t1 && temp <= t2 {
					newState = "HOT-STATE"
					newFreq = fastFreq
				} else {
					newState = "ALARM"
					newFreq = fastFreq
					inAlarm = true
				}
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

		// Case per il timeout del dispositivo ESP32.
		case <-esp32Timer.C:
			log.Println("ATTENZIONE: Dispositivo ESP32 è andato OFFLINE (timeout).")
			state.DevicesOnline["esp32"] = false
		}
	}
}

// startMqttPublisher rimane invariato.
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

// startApiServer ora accetta i due nuovi canali.
func startApiServer(commandChan chan<- RequestType, stateReqChan chan<- chan SystemState) {
	var usemock = false
	// Passa i canali corretti al costruttore del controller.
	apiController := NewController(usemock, commandChan, stateReqChan)

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
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("ERRORE: Impossibile avviare il server API: %v", err)
	}
}

func main() {
	// --- Canali per la comunicazione tra goroutine ---
	tempUpdatesChan := make(chan float64)
	// Canale per i comandi di scrittura.
	commandChan := make(chan RequestType)
	// Canale per le richieste di lettura dello stato.
	stateReqChan := make(chan chan SystemState)
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
		temp, err := strconv.ParseFloat(string(msg.Payload()), 64)
		if err != nil {
			log.Printf("ERRORE: Impossibile convertire il payload della temperatura: %v", err)
			return
		}
		tempUpdatesChan <- temp
	}

	// --- Sottoscrizione al topic della temperatura ---
	if token := client.Subscribe(tempTopic, 1, temperatureMessageHandler); token.Wait() && token.Error() != nil {
		log.Fatalf("ERRORE: Impossibile sottoscriversi al topic %s: %v", tempTopic, token.Error())
	}
	fmt.Printf("INFO: Sottoscritto con successo al topic: %s\n", tempTopic)

	// --- Avvio delle Goroutine con i canali corretti ---
	go stateManager(tempUpdatesChan, commandChan, stateReqChan, intervalUpdatesChan)
	go startApiServer(commandChan, stateReqChan)
	go startMqttPublisher(client, intervalUpdatesChan)

	fmt.Println("INFO: Tutti i servizi sono stati avviati.")
	// Blocca l'esecuzione del main per mantenere le goroutine attive.
	select {}
}
