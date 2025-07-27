package main

import (
	"encoding/binary"
	"log"
	"math"
	"net/http"
	"server/arduinoserial"
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
func stateManager(
	tempUpdatesChan <-chan float64,
	commandChan <-chan RequestType,
	stateReqChan <-chan chan SystemState,
	intervalUpdatesChan chan<- time.Duration,
	dataFromArduino <-chan arduinoserial.DataFromArduino,
	dataToArduino chan<- arduinoserial.DataToArduino,
) {
	const t1 float64 = 45.0
	const t2 float64 = 70.0
	const normalFreq time.Duration = 500 * time.Millisecond
	const fastFreq time.Duration = 100 * time.Millisecond
	const historySize = 100
	const esp32Timeout = 3 * time.Second
	const arduinoTimeout = 3 * time.Second

	var resetAlarm = false
	var inAlarm = false

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
	arduinoTimer := time.NewTimer(arduinoTimeout)
	// Ticker per inviare comandi periodici ad Arduino
	arduinoTicker := time.NewTicker(250 * time.Millisecond)
	defer arduinoTicker.Stop()

	intervalUpdatesChan <- state.SamplingInterval
	log.Println("INFO: State Manager avviato.")

	// Funzione helper per inviare comandi ad Arduino
	sendCommandToArduino := func(action int16) {
		var mode int16 = 0
		if state.OperativeMode == "MANUAL" {
			mode = 1
		}
		dataToArduino <- ArduinoCommand{
			Temperature:   int16(state.CurrentTemp),
			OperativeMode: mode,
			WindowAction:  action,
			systemState:   int16(state.SystemStatus),
		}
	}

	for {
		select {
		// Case per i comandi dall'API.
		case cmd := <-commandChan:
			var windowAction int16 = 0
			switch cmd {
			case RequestToggleMode:
				if state.OperativeMode == "AUTOMATIC" {
					state.OperativeMode = "MANUAL"
				} else {
					state.OperativeMode = "AUTOMATIC"
				}
				log.Printf("INFO: Modalità operativa cambiata a %s.", state.OperativeMode)
			case RequestOpenWindow:
				log.Println("COMANDO: Apertura finestra.")
				windowAction = 1
			case RequestCloseWindow:
				log.Println("COMANDO: Chiusura finestra.")
				windowAction = 2
			case RequestResetAlarm:
				resetAlarm = true
			}
			// Se c'è un'azione specifica (apri/chiudi), la invia subito.
			// L'aggiornamento di stato generale verrà inviato dal ticker.
			if windowAction != 0 {
				sendCommandToArduino(windowAction)
			}

		// Case per le richieste di lettura dello stato dall'API.
		case replyChan := <-stateReqChan:
			replyChan <- state

		// Case per gli aggiornamenti di temperatura da MQTT.
		case temp := <-tempUpdatesChan:
			if !state.DevicesOnline["esp32"] {
				log.Println("INFO: Dispositivo ESP32 è ora ONLINE.")
				state.DevicesOnline["esp32"] = true
			}
			esp32Timer.Reset(esp32Timeout)

			// Aggiorna statistiche temperatura
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

			// Logica di gestione dell'allarme e frequenza.
			if resetAlarm {
				resetAlarm = false
				if temp <= t2 {
					inAlarm = false
				}
			}
			newState := state.SystemStatus
			newFreq := state.SamplingInterval
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

		// Case per i dati in arrivo da Arduino.
		case data := <-dataFromArduino:
			if !state.DevicesOnline["arduino"] {
				log.Println("INFO: Dispositivo Arduino è ora ONLINE.")
				state.DevicesOnline["arduino"] = true
			}
			arduinoTimer.Reset(arduinoTimeout)
			state.WindowPosition = data.WindowPosition

		// Case per l'invio periodico ad Arduino.
		case <-arduinoTicker.C:
			sendCommandToArduino(0) // Invia stato aggiornato senza azioni specifiche

		// Timeout handlers
		case <-esp32Timer.C:
			log.Println("ATTENZIONE: Dispositivo ESP32 è andato OFFLINE (timeout).")
			state.DevicesOnline["esp32"] = false
		case <-arduinoTimer.C:
			log.Println("ATTENZIONE: Dispositivo Arduino è andato OFFLINE (timeout).")
			state.DevicesOnline["arduino"] = false
		}
	}
}

func arduinoManager(requestChan chan<- RequestType, dataFromArduino chan<- arduinoserial.DataFromArduino, dataToArduino <-chan arduinoserial.DataToArduino) {
	arduino := arduinoserial.NewArduinoReader(9600, 5*time.Second)
	if err := arduino.Connect(); err != nil {
		log.Printf("ERRORE: Impossibile connettersi ad Arduino: %v. Riprovo...", err)
		time.Sleep(5 * time.Second)
		arduinoManager(requestChan, dataFromArduino, dataToArduino)
		return
	}
	defer arduino.Disconnect()
	log.Println("INFO: Connesso ad Arduino.")

	var wasButtonPressed bool = false

	// Goroutine per la scrittura: si attiva solo quando riceve un comando.
	go func() {
		byteToSend := make([]byte, 2)

		for cmd := range dataToArduino {
			binary.LittleEndian.PutUint16(byteToSend, uint16(cmd.Temperature))
			arduino.AddDataToSend(0, byteToSend)
			binary.LittleEndian.PutUint16(byteToSend, uint16(cmd.OperativeMode))
			arduino.AddDataToSend(1, byteToSend)
			binary.LittleEndian.PutUint16(byteToSend, uint16(cmd.WindowAction))
			arduino.AddDataToSend(2, byteToSend)

			if err := arduino.WriteData(); err != nil {
				log.Printf("ERRORE: Impossibile inviare dati ad Arduino: %v", err)
			}
		}
	}()

	// Loop principale per la lettura continua da Arduino
	for {
		vars, _, _, err := arduino.ReadData()
		if err != nil {
			//timeout, is not critical
			continue
		}

		if len(vars) < 2 {
			log.Println("WARN: Ricevuto pacchetto incompleto da Arduino.")
			continue
		}

		buttonState, ok1 := vars[0].Data.(int16)
		windowPos, ok2 := vars[1].Data.(int16)
		if !ok1 || !ok2 {
			log.Println("ERRORE: Dati da Arduino non validi o tipo inatteso.")
			continue
		}

		isButtonPressed := (buttonState == 1)
		// Rileva il fronte di salita del pulsante per inviare un solo comando
		if isButtonPressed && !wasButtonPressed {
			log.Println("INFO: Pressione pulsante rilevata, invio comando ToggleMode.")
			requestChan <- ToggleMode
		}
		wasButtonPressed = isButtonPressed

		dataFromArduino <- arduinoserial.DataFromArduino{WindowPosition: int(windowPos)}
	}
}

func mqttPublisher(client MQTT.Client, IntervalUpdatesChan <-chan time.Duration) {
	const configTopic = "esp32/config/interval"
	log.Println("INFO: Publisher MQTT avviato.")
	for interval := range IntervalUpdatesChan {
		intervalPayload := strconv.FormatInt(interval.Milliseconds(), 10)
		token := client.Publish(configTopic, 1, false, intervalPayload)
		token.Wait()
	}
}

// startApiServer rimane invariato.
func startApiServer(commandChan chan<- RequestType, stateReqChan chan<- chan SystemState) {
	useMock := false
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
	log.Println("INFO: API in ascolto su :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("ERRORE: Impossibile avviare il server API: %v", err)
	}
}

func main() {
	// --- Canali per la comunicazione tra goroutine ---
	tempUpdatesChan := make(chan float64)
	RequestChan := make(chan RequestType)
	stateReqChan := make(chan chan SystemState)
	intervalUpdatesChan := make(chan time.Duration)
	dataFromArduinoChan := make(chan arduinoserial.DataFromArduino)
	dataToArduinoChan := make(chan arduinoserial.DataToArduino)

	// --- Configurazione e connessione MQTT ---
	const broker = "tcp://localhost:1883"
	const tempTopic = "esp32/data/temperature"
	opts := MQTT.NewClientOptions().AddBroker(broker).SetClientID("go-server-logic")
	client := MQTT.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		log.Fatalf("ERRORE: Impossibile connettersi al broker MQTT: %v", token.Error())
	}
	log.Println("INFO: Connesso al broker MQTT.")

	// --- Handler per i messaggi di temperatura ---
	var temperatureMessageHandler MQTT.MessageHandler = func(client MQTT.Client, msg MQTT.Message) {
		temp, err := strconv.ParseFloat(string(msg.Payload()), 64)
		if err == nil {
			tempUpdatesChan <- temp
		}
	}
	if token := client.Subscribe(tempTopic, 1, temperatureMessageHandler); token.Wait() && token.Error() != nil {
		log.Fatalf("ERRORE: Impossibile sottoscriversi al topic %s: %v", tempTopic, token.Error())
	}
	log.Printf("INFO: Sottoscritto con successo al topic: %s\n", tempTopic)

	// --- Avvio delle Goroutine ---
	go stateManager(tempUpdatesChan, RequestChan, stateReqChan, intervalUpdatesChan, dataFromArduinoChan, dataToArduinoChan)
	go startApiServer(RequestChan, stateReqChan)
	go mqttPublisher(client, intervalUpdatesChan)
	go arduinoManager(RequestChan, dataFromArduinoChan, dataToArduinoChan)

	log.Println("INFO: Tutti i servizi sono stati avviati.")
	select {}
}
