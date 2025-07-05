package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	// Crea un nuovo reader con baudrate 9600 e timeout di 1 secondo.
	arduino := NewArduinoReader(9600, 1*time.Second)

	// Connettiti ad Arduino.
	if err := arduino.Connect(); err != nil {
		log.Fatalf("Impossibile connettersi: %v", err)
	}
	// Assicura che la connessione venga chiusa alla fine del programma.
	defer arduino.Disconnect()

	// Canale per segnalare la fine alle goroutine.
	done := make(chan struct{})

	// --- Goroutine per la SCRITTURA ---
	go func() {
		// Un Ticker è un modo efficiente per eseguire un'azione a intervalli regolari.
		ticker := time.NewTicker(1 * time.Second) // Scrive ogni secondo
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				// Questo blocco viene eseguito a ogni "tick" del ticker.
				log.Println("SCRITTURA: Invio dati ad Arduino...")
				if err := arduino.WriteData(25, 1); err != nil {
					log.Printf("Errore durante la scrittura: %v", err)
				}
			case <-done:
				// Se il canale 'done' viene chiuso, la goroutine termina.
				log.Println("Goroutine di scrittura terminata.")
				return
			}
		}
	}()

	// --- Goroutine per la LETTURA (eseguita nel main) ---
	go func() {
		for {
			select {
			case <-done:
				// Se il canale 'done' viene chiuso, la goroutine termina.
				log.Println("Goroutine di lettura terminata.")
				return
			default:
				// Esegue la lettura
				vars, debugs, _, err := arduino.ReadData()
				if err != nil {
					// Ignoriamo gli errori di timeout che sono normali se non ci sono dati
					continue
				}

				for _, v := range vars {
					fmt.Printf("LETTURA -> VAR Ricevuto -> ID: %d, Tipo: %d, Valore: %v\n", v.ID, v.VarType, v.Data)
				}
				for _, d := range debugs {
					fmt.Printf("LETTURA -> DEBUG Ricevuto -> ID: %d, Tipo: %d, Messaggio: %v\n", d.ID, d.VarType, d.Data)
				}
			}
		}
	}()

	// --- Gestione della chiusura pulita ---
	// Attende un segnale di interruzione (es. Ctrl+C)
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan // Il programma si blocca qui finché non riceve il segnale

	log.Println("Segnale di interruzione ricevuto. Chiusura in corso...")
	close(done) // Chiude il canale 'done' per segnalare a tutte le goroutine di terminare.

	// Attendi un istante per permettere alle goroutine di terminare prima che defer venga eseguito.
	time.Sleep(100 * time.Millisecond)
}
