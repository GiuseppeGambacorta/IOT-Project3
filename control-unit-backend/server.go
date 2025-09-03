package main

import (
	"context"
	"log"
	"math"
	"os"
	"os/signal"
	"server/arduinoserial"
	"server/mqtt"
	"server/system"
	"server/webserver"
	"strconv"
	"sync"
	"syscall"
	"time"

	MQTT "github.com/eclipse/paho.mqtt.golang"
)

func systemManager(
	ctx context.Context,
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

	var tempHistory = make([]float64, 0, system.MaxTemperatureBuffer)

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

loop:
	for {
		select {
		case temp := <-tempUpdatesChan:
			esp32TimeoutTimer.Reset(esp32TimeoutTime)
			if !actualSystemState.DevicesOnline["esp32"] {
				log.Println("INFO: Dispositivo ESP32 è ora ONLINE.")
				actualSystemState.DevicesOnline["esp32"] = true
			}

			tempHistory = system.ManageTemperature(temp, tempHistory, &actualSystemState)

			system.ManageSystemLogic(
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
				system.ToggleActualMode(&actualSystemState)
			case system.OpenWindow:
				if actualSystemState.OperativeMode == system.Manual {
					windowManualCommand = system.CmdCloseWindow
				} else {
					windowManualCommand = system.NoCommand
				}
			case system.CloseWindow:
				if actualSystemState.OperativeMode == system.Manual {
					windowManualCommand = system.CmdCloseWindow
				} else {
					windowManualCommand = system.NoCommand
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
				system.ToggleActualMode(&actualSystemState)
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

		case <-ctx.Done():
			log.Println("System Manager : Shutdown")
			break loop
		}

	}
}
func main() {
	useMockApi := false
	var wg sync.WaitGroup

	startGoroutine := func(fn func()) {
		wg.Add(1)
		go func() {
			defer wg.Done()
			fn()
		}()
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Goroutine per ascoltare SIGINT/SIGTERM
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh
		cancel()
	}()

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
		log.Println(err)
		goto Error
	}

	startGoroutine(func() {
		systemManager(ctx, intervalUpdatesChan, tempUpdatesChan, stateRequestChan, commandRequestChan, dataToArduinoChan, dataFromArduinoChan)
	})

	startGoroutine(func() { mqtt.MqttPublishInterval(ctx, client, intervalUpdatesChan) })

	startGoroutine(func() { webserver.ApiServer(ctx, useMockApi, commandRequestChan, stateRequestChan) })

	startGoroutine(func() { arduinoserial.ManageArduino(ctx, dataFromArduinoChan, dataToArduinoChan) })

	log.Println("INFO: Tutti i servizi sono stati avviati.")

	<-ctx.Done() // Blocca finché non riceve segnale di chiusura
	log.Println("Shutdown in corso")
	wg.Wait()
	log.Println("Shutdown completato.")

Error:
}
