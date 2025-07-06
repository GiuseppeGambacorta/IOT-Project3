package main

import (
	"fmt"
	"log"
	"time"
)

type WindowManagerState int

const (
	AUTOMATIC WindowManagerState = iota // Assegna 0
	MANUAL                              // Assegna 1
)

func main() {

	arduino := NewArduinoReader(9600, 5*time.Second)
	var actualState WindowManagerState = AUTOMATIC
	var oldbuttonState bool

	if err := arduino.Connect(); err != nil {
		log.Fatalf("Impossibile connettersi: %v", err)
	}

	defer arduino.Disconnect()
	var num int16 = 0
	for {
		vars, _, _, err := arduino.ReadData()
		if err != nil {
			fmt.Printf("Errore durante la lettura dei dati: %v\n", err)
			//continue
		}

		if len(vars) < 2 {
			fmt.Println("Nessuna variabile ricevuta.")
			//continue
		}

		var isButtonPressed = vars[0].Data.(int16) == 1
		var windowPosition = vars[1].Data.(int16)

		if !isButtonPressed {
			oldbuttonState = false
		}

		fmt.Printf("pulsante: %v\n", isButtonPressed)

		if isButtonPressed && !oldbuttonState {
			if actualState == AUTOMATIC {
				actualState = MANUAL
				oldbuttonState = isButtonPressed
				fmt.Println("Modalità manuale attivata.")
			} else {
				actualState = AUTOMATIC
				oldbuttonState = isButtonPressed
				fmt.Println("Modalità automatica attivata.")
			}
		}

		fmt.Printf("Stato attuale: %v, Posizione finestra: %d\n", actualState, windowPosition)

		time.Sleep(250 * time.Millisecond)

		arduino.addDataToSend(0, num)
		num = num + 1
		if num > 90 {
			num = 0
		}
		arduino.addDataToSend(1, int16(actualState))
		arduino.addDataToSend(2, windowPosition)
		arduino.WriteData()
	}
}
