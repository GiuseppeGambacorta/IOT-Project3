package main

import (
	"fmt"
	"log"
	"time"
)

type WindowManagerState int

// Definiamo i valori possibili per lo stato.
const (
	AUTOMATIC WindowManagerState = iota // Assegna 0
	MANUAL                              // Assegna 1
)

func main() {
	// Crea un nuovo reader con baudrate 9600 e timeout di 1 secondo.
	arduino := NewArduinoReader(9600, 5*time.Second)
	var actualState WindowManagerState = AUTOMATIC
	var oldbuttonState bool
	// Connettiti ad Arduino.
	if err := arduino.Connect(); err != nil {
		log.Fatalf("Impossibile connettersi: %v", err)
	}
	// Assicura che la connessione venga chiusa alla fine.
	defer arduino.Disconnect()

	// Loop per leggere i dati.
	for {
		vars, _, _, err := arduino.ReadData()
		if err != nil {
			fmt.Printf("Errore durante la lettura dei dati: %v\n", err)
			//continue
		}

		if len(vars) < 3 {
			fmt.Println("Nessuna variabile ricevuta.")
			//continue
		}

		var isButtonPressed = vars[0].Data.(int16) == 1
		var windowPosition = vars[1].Data.(int16)
		var statochearriva = vars[2].Data.(int16)

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
		fmt.Printf("Stato che arriva: %d\n", statochearriva)

		// Attendi un po' prima della prossima lettura.
		time.Sleep(250 * time.Millisecond)
		arduino.WriteData(50, 1)
		arduino.WriteData(int16(actualState), 0)
	}
}
