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

type Channels struct {
	IntervalUpdatesChan chan time.Duration
	TempUpdatesChan     chan float64
	CommandRequestChan  chan system.RequestType
	StateRequestChan    chan chan system.SystemState
	DataFromArduinoChan chan arduinoserial.DataFromArduino
	DataToArduinoChan   chan arduinoserial.DataToArduino
}

func systemManager(
	ctx context.Context,
	ch Channels,
) {
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
		case temp := <-ch.TempUpdatesChan:
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
				ch.IntervalUpdatesChan,
				&tooHotEnteredAt,
				tooHotMaxDuration,
			)

		case stateRequest := <-ch.StateRequestChan:
			stateRequest <- actualSystemState

		case commandRequest := <-ch.CommandRequestChan:
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
						actualSystemState.Status = system.Normal
					}
				}
			default:
				log.Println("Comando sconosciuto")
			}

		case data := <-ch.DataFromArduinoChan:
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
				windowManualCommand = 0
				select {
				case ch.DataToArduinoChan <- newData:
				default:
					<-ch.DataToArduinoChan
					ch.DataToArduinoChan <- newData
					log.Println("WARN: Buffer dati per Arduino pieno. sostituto valore.")
				}
			}

		case <-esp32TimeoutTimer.C:
			esp32TimeoutTimer.Reset(esp32TimeoutTime)
			if actualSystemState.DevicesOnline["esp32"] {
				log.Println("ATTENZIONE: Dispositivo ESP32 è andato OFFLINE (timeout).")
				actualSystemState.DevicesOnline["esp32"] = false
			} else {
				ch.IntervalUpdatesChan <- actualSystemState.SamplingInterval
			}

		case <-ctx.Done():
			log.Println("System Manager : Shutdown")
			break loop
		}
	}
}

func main() {
	useMockApi := true
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

	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh
		cancel()
	}()

	// --- Canali ---
	ch := Channels{
		IntervalUpdatesChan: make(chan time.Duration),
		TempUpdatesChan:     make(chan float64),
		CommandRequestChan:  make(chan system.RequestType),
		StateRequestChan:    make(chan chan system.SystemState),
		DataFromArduinoChan: make(chan arduinoserial.DataFromArduino, 20),
		DataToArduinoChan:   make(chan arduinoserial.DataToArduino, 1),
	}

	// --- MQTT ---
	const broker = "tcp://localhost:1883"
	const cliendID = "iot-server"
	const tempTopic = "esp32/data/temperature"

	var temperatureMessageHandler MQTT.MessageHandler = func(client MQTT.Client, msg MQTT.Message) {
		temp, err := strconv.ParseFloat(string(msg.Payload()), 64)
		if err == nil {
			ch.TempUpdatesChan <- temp
		} else {
			log.Println("errore lettura temperatura")
		}
	}

	client, err := mqtt.ConfigureClient(broker, cliendID,
		func(c MQTT.Client) {
			if token := c.Subscribe(tempTopic, 1, temperatureMessageHandler); token.Wait() && token.Error() != nil {
				log.Printf("MQTT: errore nella risottoscrizione: %v", token.Error())
			}
		})

	if err != nil {
		log.Println(err)
		goto Error
	}

	startGoroutine(func() {
		systemManager(ctx, ch)
	})

	startGoroutine(func() { mqtt.MqttPublishInterval(ctx, client, ch.IntervalUpdatesChan) })

	startGoroutine(func() { webserver.ApiServer(ctx, useMockApi, ch.CommandRequestChan, ch.StateRequestChan) })

	startGoroutine(func() { arduinoserial.ManageArduino(ctx, ch.DataFromArduinoChan, ch.DataToArduinoChan) })

	log.Println("INFO: Tutti i servizi sono stati avviati.")

	<-ctx.Done()
	log.Println("Shutdown in corso")
	wg.Wait()
	log.Println("Shutdown completato.")

Error:
}
