package main

import (
	"fmt"
	"log"
	"time"
)

func main() {
	// Crea un nuovo reader con baudrate 9600 e timeout di 1 secondo.
	arduino := NewArduinoReader(9600, 1*time.Second)

	// Connettiti ad Arduino.
	if err := arduino.Connect(); err != nil {
		log.Fatalf("Impossibile connettersi: %v", err)
	}
	// Assicura che la connessione venga chiusa alla fine.
	defer arduino.Disconnect()

	// Loop per leggere i dati.
	for {
		vars, debugs, _, err := arduino.ReadData()
		if err != nil {
			// Se c'è un errore di timeout, è normale se Arduino non invia dati.
			// Puoi ignorarlo o gestirlo come preferisci.
			// log.Printf("Errore di lettura: %v", err)
			continue
		}

		// Stampa le variabili ricevute
		for _, v := range vars {
			fmt.Printf("VAR Ricevuto -> ID: %d, Tipo: %d, Valore: %v\n", v.ID, v.VarType, v.Data)
		}

		// Stampa i messaggi di debug ricevuti
		for _, d := range debugs {
			fmt.Printf("DEBUG Ricevuto -> ID: %d, Tipo: %d, Messaggio: %v\n", d.ID, d.VarType, d.Data)
		}

		// Attendi un po' prima della prossima lettura.
		//time.Sleep(500 * time.Millisecond)
		arduino.WriteData(1, 0)
		//arduino.WriteData(1, 1)
	}
}
