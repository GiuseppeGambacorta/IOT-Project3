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
	stateReqChan <-chan chan system.System,
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

	systemState := system.System{
		Status:           system.Normal,
		SamplingInterval: normalFreq,
		OperativeMode:    system.Automatic,
		MinTemp:          math.Inf(1),
		MaxTemp:          math.Inf(-1),
		DevicesOnline: map[system.DeviceName]bool{
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

	intervalUpdatesChan <- systemState.SamplingInterval
	log.Println("INFO: State Manager avviato.")

	// Funzione helper per inviare comandi ad Arduino
	sendCommandToArduino := func(action int16) {
		var mode int16 = 0
		if systemState.OperativeMode == system.Manual {
			mode = 1
		}
		dataToArduino <- arduinoserial.DataToArduino{
			Temperature:   int16(systemState.CurrentTemp),
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
				if systemState.OperativeMode == system.Automatic {
					systemState.OperativeMode = system.Manual
				} else {
					systemState.OperativeMode = system.Automatic
				}
				log.Printf("INFO: Modalità operativa cambiata a %s.", systemState.OperativeMode)
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
			replyChan <- systemState

		// Case per gli aggiornamenti di temperatura da MQTT.
		case temp := <-tempUpdatesChan:
			if !systemState.DevicesOnline["esp32"] {
				log.Println("INFO: Dispositivo ESP32 è ora ONLINE.")
				systemState.DevicesOnline["esp32"] = true
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
			systemState.CurrentTemp = temp
			systemState.AverageTemp = sum / float64(len(tempHistory))
			systemState.MinTemp = min
			systemState.MaxTemp = max

			// Logica di gestione dell'allarme e frequenza.
			if resetAlarm {
				resetAlarm = false
				if temp <= t2 {
					inAlarm = false
				}
			}
			newStatus := systemState.Status
			newFreq := systemState.SamplingInterval
			if !inAlarm {
				if temp <= t1 {
					newStatus = system.Normal
					newFreq = normalFreq
				} else if temp > t1 && temp <= t2 {
					newStatus = system.HotState
					newFreq = fastFreq
				} else {
					newStatus = system.Alarm
					newFreq = fastFreq
					inAlarm = true
				}
			}
			if newStatus != systemState.Status {
				log.Printf("ATTENZIONE: Cambio di stato -> %s (Temp: %.1f°C)", newStatus, temp)
				systemState.Status = newStatus

			}
			if newFreq != systemState.SamplingInterval {
				log.Printf("INFO: Frequenza di campionamento cambiata a %v", newFreq)
				systemState.SamplingInterval = newFreq
				intervalUpdatesChan <- systemState.SamplingInterval
			}

		// Case per i dati in arrivo da Arduino.
		case data := <-dataFromArduino:
			if !systemState.DevicesOnline["arduino"] {
				log.Println("INFO: Dispositivo Arduino è ora ONLINE.")
				systemState.DevicesOnline["arduino"] = true
			}
			arduinoTimer.Reset(arduinoTimeout)
			systemState.WindowPosition = data.WindowPosition

		// Case per l'invio periodico ad Arduino.
		case <-arduinoTicker.C:
			sendCommandToArduino(0) // Invia stato aggiornato senza azioni specifiche

		// Timeout handlers
		case <-esp32Timer.C:
			log.Println("ATTENZIONE: Dispositivo ESP32 è andato OFFLINE (timeout).")
			systemState.DevicesOnline["esp32"] = false
		case <-arduinoTimer.C:
			log.Println("ATTENZIONE: Dispositivo Arduino è andato OFFLINE (timeout).")
			systemState.DevicesOnline["arduino"] = false
		}
	}
}

func main() {
	// --- Canali per la comunicazione tra goroutine ---
	tempUpdatesChan := make(chan float64)
	dataFromArduinoChan := make(chan arduinoserial.DataFromArduino, 20)
	dataToArduinoChan := make(chan arduinoserial.DataToArduino)
	RequestChan := make(chan system.RequestType)
	stateReqChan := make(chan chan system.System)
	intervalUpdatesChan := make(chan time.Duration)

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
		log.Panicf("ERRORE:Impossibile sottoscrivere il topic %s", tempTopic)
	}

	// --- Avvio delle Goroutine ---
	go stateManager(tempUpdatesChan, RequestChan, stateReqChan, intervalUpdatesChan, dataFromArduinoChan, dataToArduinoChan)
	go webserver.ApiServer(RequestChan, stateReqChan)
	go mqtt.MqttPublisher(client, intervalUpdatesChan)
	go arduinoserial.ManageArduino(RequestChan, dataFromArduinoChan, dataToArduinoChan)

	log.Println("INFO: Tutti i servizi sono stati avviati.")
	select {}
}
