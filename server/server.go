package main

import (
	"log"
	"math"
	"strconv"

	"server/arduinoserial"
	"server/mqtt"
	"server/system"
	"server/webserver"
	"time"

	MQTT "github.com/eclipse/paho.mqtt.golang"
)

// stateManager è la goroutine centrale che gestisce lo stato dell'applicazione.
func stateManager(
	tempUpdatesChan <-chan float64,
	commandChan <-chan system.RequestType,
	stateReqChan <-chan chan system.SystemState,
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

	state := system.SystemState{
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
		dataToArduino <- arduinoserial.DataToArduino{
			Temperature:   int16(state.CurrentTemp),
			OperativeMode: mode,
			WindowAction:  action,
			SystemState:   0,
		}
	}

	for {
		select {
		// Case per i comandi dall'API.
		case cmd := <-commandChan:
			var windowAction int16 = 0
			switch cmd {
			case system.ToggleMode:
				if state.OperativeMode == "AUTOMATIC" {
					state.OperativeMode = "MANUAL"
				} else {
					state.OperativeMode = "AUTOMATIC"
				}
				log.Printf("INFO: Modalità operativa cambiata a %s.", state.OperativeMode)
			case system.OpenWindow:
				log.Println("COMANDO: Apertura finestra.")
				windowAction = 1
			case system.CloseWindow:
				log.Println("COMANDO: Chiusura finestra.")
				windowAction = 2
			case system.ResetAlarm:
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

func main() {
	// --- Canali per la comunicazione tra goroutine ---
	tempUpdatesChan := make(chan float64)
	RequestChan := make(chan system.RequestType)
	stateReqChan := make(chan chan system.SystemState)
	intervalUpdatesChan := make(chan time.Duration)
	dataFromArduinoChan := make(chan arduinoserial.DataFromArduino)
	dataToArduinoChan := make(chan arduinoserial.DataToArduino)

	// --- Configurazione e connessione MQTT ---
	const broker = "tcp://localhost:1883"
	const tempTopic = "esp32/data/temperature"
	client := mqtt.ConfigureClient(broker)

	// --- Handler per i messaggi di temperatura ---
	var temperatureMessageHandler MQTT.MessageHandler = func(client MQTT.Client, msg MQTT.Message) {
		temp, err := strconv.ParseFloat(string(msg.Payload()), 64)
		if err == nil {
			tempUpdatesChan <- temp
		}
	}

	err := mqtt.SubscribeToTopic(client, temperatureMessageHandler, tempTopic)
	if err != nil {
		log.Panic("ERRORE:Impossibile sottoscrivere il topic %s", tempTopic)
	}

	// --- Avvio delle Goroutine ---
	go stateManager(tempUpdatesChan, RequestChan, stateReqChan, intervalUpdatesChan, dataFromArduinoChan, dataToArduinoChan)
	go webserver.ApiServer(RequestChan, stateReqChan)
	go mqtt.MqttPublisher(client, intervalUpdatesChan)
	go arduinoserial.ManageArduino(RequestChan, dataFromArduinoChan, dataToArduinoChan)

	log.Println("INFO: Tutti i servizi sono stati avviati.")
	select {}
}
