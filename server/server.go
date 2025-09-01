package main

import (
	"log"
	"math"
	"server/arduinoserial"
	"server/mqtt"
	"server/system"
	"server/webserver"
	"strconv"
	"time"

	MQTT "github.com/eclipse/paho.mqtt.golang"
)

const (
	MaxTemperatureBuffer = 100
)

// stateManager è la goroutine centrale che gestisce lo stato dell'applicazione.
/*
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
				log.Printf("INFO: Modalità operativa cambiata a %s.", systemState.OperativeMode.String())
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
*/

func toggleMode(actualSystemState *system.SystemState) {
	if actualSystemState.OperativeMode == system.Manual {
		actualSystemState.OperativeMode = system.Automatic
	} else {
		actualSystemState.OperativeMode = system.Manual
	}
	actualSystemState.OperativeModeString = actualSystemState.OperativeMode.String()
	log.Println("Modalita attuale: " + actualSystemState.OperativeModeString)
}

func manageTemperature(temp float64, tempHistory []float64, actualSystemState *system.SystemState) []float64 {
	tempHistory = append(tempHistory, temp)
	if len(tempHistory) > MaxTemperatureBuffer {
		tempHistory = tempHistory[1:]
	}
	var sum float64
	for _, t := range tempHistory {
		sum += t
		if t < actualSystemState.MinTemp {
			actualSystemState.MinTemp = t
		}
		if t > actualSystemState.MaxTemp {
			actualSystemState.MaxTemp = t
		}
	}
	actualSystemState.CurrentTemp = temp
	actualSystemState.AverageTemp = sum / float64(len(tempHistory))
	return tempHistory
}

func manageMotorPosition(actualSystemState *system.SystemState, threshold1, threshold2 float64) {
	switch actualSystemState.Status {
	case system.Alarm, system.Too_hot:
		actualSystemState.CommandWindowPosition = 90
	case system.Hot:
		actualSystemState.CommandWindowPosition = system.Degree(
			(actualSystemState.CurrentTemp - threshold1) * (threshold2 / (threshold2 - threshold1)),
		)
	default:
		actualSystemState.CommandWindowPosition = 0
	}
}

func manageSystemLogic(
	actualSystemState *system.SystemState,
	threshold1, threshold2 float64,
	normalFreq, fastFreq time.Duration,
	intervalUpdatesChan chan<- time.Duration,
	tooHotEnteredAt *time.Time,
	tooHotMaxDuration time.Duration) {

	oldStatus := actualSystemState.Status
	oldFreq := actualSystemState.SamplingInterval
	now := time.Now()

	if actualSystemState.Status != system.Alarm {
		if actualSystemState.CurrentTemp <= threshold1 {
			actualSystemState.Status = system.Normal
			actualSystemState.SamplingInterval = normalFreq
			*tooHotEnteredAt = time.Time{}
		} else if actualSystemState.CurrentTemp > threshold1 && actualSystemState.CurrentTemp <= threshold2 {
			actualSystemState.Status = system.Hot
			actualSystemState.SamplingInterval = fastFreq
			*tooHotEnteredAt = time.Time{}
		} else {
			actualSystemState.Status = system.Too_hot
			actualSystemState.SamplingInterval = fastFreq
			if tooHotEnteredAt.IsZero() {
				*tooHotEnteredAt = now
			}
			if !tooHotEnteredAt.IsZero() && now.Sub(*tooHotEnteredAt) > tooHotMaxDuration {
				actualSystemState.Status = system.Alarm
				log.Println("ALLARME: Temperatura troppo alta per troppo tempo! Stato -> Alarm")
				*tooHotEnteredAt = time.Time{}
			}
		}
	}

	manageMotorPosition(actualSystemState, threshold1, threshold2)
	actualSystemState.StatusString = actualSystemState.Status.String()
	actualSystemState.OperativeModeString = actualSystemState.OperativeMode.String()
	if actualSystemState.Status != oldStatus {
		log.Printf("ATTENZIONE: Cambio di stato -> %s (Temp: %.1f°C)", actualSystemState.Status.String(), actualSystemState.CurrentTemp)
	}
	if actualSystemState.SamplingInterval != oldFreq {
		log.Printf("INFO: Frequenza di campionamento cambiata a %v", actualSystemState.SamplingInterval)
		intervalUpdatesChan <- actualSystemState.SamplingInterval
	}
}

func systemManager(
	intervalUpdatesChan chan<- time.Duration,
	tempUpdatesChan <-chan float64,
	stateRequestChan <-chan chan system.SystemState,
	commandRequestChan <-chan system.RequestType,
	dataToArduinoChan chan arduinoserial.DataToArduino,
	dataFromArduinoChan <-chan arduinoserial.DataFromArduino) {

	const (
		normalFreq time.Duration = 500 * time.Millisecond
		fastFreq   time.Duration = 100 * time.Millisecond
		threshold1 float64       = 30
		threshold2 float64       = 70

		esp32TimeoutTime  = 2 * time.Second
		tooHotMaxDuration = 10 * time.Second

		arduinoSerialFreq = 250 * time.Millisecond
	)

	esp32TimeoutTimer := time.NewTimer(esp32TimeoutTime)
	arduinoTimer := time.NewTimer(arduinoSerialFreq)
	var tooHotEnteredAt time.Time

	var tempHistory = make([]float64, 0, MaxTemperatureBuffer)

	windowManualCommand := 0

	actualSystemState := system.SystemState{
		Status:              system.Normal,
		StatusString:        system.Normal.String(),
		SamplingInterval:    normalFreq,
		OperativeMode:       system.Automatic,
		OperativeModeString: system.Automatic.String(),
		CurrentTemp:         0,
		AverageTemp:         0,
		MinTemp:             math.Inf(1),
		MaxTemp:             math.Inf(-1),
		WindowPosition:      0,
		DevicesOnline: map[system.DeviceName]bool{
			"server":  true,
			"esp32":   false,
			"arduino": false,
		},
	}

	for {
		select {
		case temp := <-tempUpdatesChan:
			esp32TimeoutTimer.Reset(esp32TimeoutTime)
			if !actualSystemState.DevicesOnline["esp32"] {
				log.Println("INFO: Dispositivo ESP32 è ora ONLINE.")
				actualSystemState.DevicesOnline["esp32"] = true
			}

			tempHistory = manageTemperature(temp, tempHistory, &actualSystemState)

			manageSystemLogic(
				&actualSystemState,
				threshold1, threshold2,
				normalFreq, fastFreq,
				intervalUpdatesChan,
				&tooHotEnteredAt,
				tooHotMaxDuration,
			)

		case stateRequest := <-stateRequestChan:
			stateRequest <- actualSystemState

		case commandRequest := <-commandRequestChan:
			switch commandRequest {
			case system.ToggleMode:
				toggleMode(&actualSystemState)
			case system.OpenWindow:
				if actualSystemState.OperativeMode == system.Manual {
					windowManualCommand = 1
				} else {
					windowManualCommand = 0
				}
			case system.CloseWindow:
				if actualSystemState.OperativeMode == system.Manual {
					windowManualCommand = 2
				} else {
					windowManualCommand = 0
				}
			case system.ResetAlarm:
				if actualSystemState.Status == system.Alarm {
					if actualSystemState.CurrentTemp < threshold2 {
						actualSystemState.Status = system.Normal // non perfetto, dovrei decidere in base alla temperatura, lo faccio fare quando la temperatura viene aggiornata
					}
				}
			default:
				log.Println("Comando sconosciuto")
			}

		case data := <-dataFromArduinoChan:
			actualSystemState.WindowPosition = data.WindowPosition
			if data.ButtonPressed {
				toggleMode(&actualSystemState)
			}
			actualSystemState.DevicesOnline["arduino"] = true

		case <-arduinoTimer.C:
			arduinoTimer.Reset(arduinoSerialFreq)
			if actualSystemState.DevicesOnline["arduino"] {
				newData := arduinoserial.DataToArduino{
					Temperature:          int(actualSystemState.CurrentTemp),
					OperativeMode:        int(actualSystemState.OperativeMode),
					WindowAction:         windowManualCommand,
					SystemState:          int(actualSystemState.Status),
					SystemWindowPosition: actualSystemState.CommandWindowPosition,
				}
				windowManualCommand = 0 // resetto sempre
				select {
				case dataToArduinoChan <- newData:

				default:
					<-dataToArduinoChan
					dataToArduinoChan <- newData
					log.Println("WARN: Buffer dati per Arduino pieno. sostituto valore.")
				}
			}

		case <-esp32TimeoutTimer.C:
			esp32TimeoutTimer.Reset(esp32TimeoutTime)
			if actualSystemState.DevicesOnline["esp32"] {
				log.Println("ATTENZIONE: Dispositivo ESP32 è andato OFFLINE (timeout).")
				actualSystemState.DevicesOnline["esp32"] = false
			} else {
				intervalUpdatesChan <- actualSystemState.SamplingInterval // aggiorno la frequenza, in modo che esp32 si possa ricollegare
			}

		}

	}
}
func main() {
	/*
		useMockApi :=
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
			goto Error
		}

		// --- Avvio delle Goroutine ---
		go stateManager(tempUpdatesChan, RequestChan, stateReqChan, intervalUpdatesChan, dataFromArduinoChan, dataToArduinoChan)
		go webserver.ApiServer(useMockApi, RequestChan, stateReqChan)
		go mqtt.MqttPublisher(client, intervalUpdatesChan)
		go arduinoserial.MansageArduino(RequestChan, dataFromArduinoChan, dataToArduinoChan)
	*/

	useMockApi := false

	intervalUpdatesChan := make(chan time.Duration)
	tempUpdatesChan := make(chan float64)
	commandRequestChan := make(chan system.RequestType)
	stateRequestChan := make(chan chan system.SystemState)

	dataFromArduinoChan := make(chan arduinoserial.DataFromArduino, 20) //buffered chan, se e piena, faccio lo shift dei dati
	dataToArduinoChan := make(chan arduinoserial.DataToArduino, 1)      // buffered chan, se e piena, scarto il vecchio comando e metto quello nuovo

	// --- Configurazione e connessione MQTT ---
	const broker = "tcp://localhost:1883"
	const cliendID = "iot-server"
	const tempTopic = "esp32/data/temperature"

	// --- Handler per i messaggi di temperatura ---
	var temperatureMessageHandler MQTT.MessageHandler = func(client MQTT.Client, msg MQTT.Message) {
		temp, err := strconv.ParseFloat(string(msg.Payload()), 64)
		if err == nil {

			tempUpdatesChan <- temp
		} else {
			log.Println("errore lettura temperatura")
		}
	}

	client, err := mqtt.ConfigureClient(broker, cliendID,
		func(c MQTT.Client) {
			if token := c.Subscribe("esp32/data/temperature", 1, temperatureMessageHandler); token.Wait() && token.Error() != nil {
				log.Printf("MQTT: errore nella risottoscrizione: %v", token.Error())
			}
		})

	if err != nil {
		goto Error
	}

	go systemManager(intervalUpdatesChan, tempUpdatesChan, stateRequestChan, commandRequestChan, dataToArduinoChan, dataFromArduinoChan)
	go mqtt.MqttPublishInterval(client, intervalUpdatesChan)
	go webserver.ApiServer(useMockApi, commandRequestChan, stateRequestChan)
	go arduinoserial.ManageArduino(dataFromArduinoChan, dataToArduinoChan)
	log.Println("INFO: Tutti i servizi sono stati avviati.")

	select {}

Error:
	log.Println(err)
}
