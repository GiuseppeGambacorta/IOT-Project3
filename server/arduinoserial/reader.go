package main

import (
	"fmt"
	"time"

	"go.bug.st/serial"
)

type ArduinoReader struct {
	portName string
	baudrate int
	timeout  time.Duration
	conn     serial.Port
	protocol *Protocol

	Variables []*DataHeader
	Debugs    []*DataHeader
	Events    []*DataHeader
}

func NewArduinoReader(baudrate int, timeout time.Duration) *ArduinoReader {
	return &ArduinoReader{
		baudrate: baudrate,
		timeout:  timeout,
	}
}

func (ar *ArduinoReader) findArduinoPort() (string, error) {
	ports, err := serial.GetPortsList()
	if err != nil {
		return "", err
	}
	if len(ports) == 0 {
		return "", fmt.Errorf("nessuna porta seriale trovata")
	}
	return ports[0], nil
}

func (ar *ArduinoReader) Connect() error {
	var err error
	ar.portName, err = ar.findArduinoPort()
	if err != nil {
		return fmt.Errorf("errore nella ricerca della porta: %w", err)
	}

	mode := &serial.Mode{
		BaudRate: ar.baudrate,
	}
	ar.conn, err = serial.Open(ar.portName, mode)
	if err != nil {
		return fmt.Errorf("impossibile aprire la porta %s: %w", ar.portName, err)
	}
	ar.conn.SetReadTimeout(ar.timeout)

	ar.protocol = NewProtocol(ar.conn)
	fmt.Printf("Connesso ad Arduino su %s a %d baud.\n", ar.portName, ar.baudrate)

	if err := ar.protocol.Handshake(); err != nil {
		ar.conn.Close()
		return fmt.Errorf("handshake fallito: %w", err)
	}

	return nil
}

func (ar *ArduinoReader) Disconnect() {
	if ar.conn != nil {
		ar.conn.Close()
		fmt.Println("Connessione chiusa.")
	}
}

func (ar *ArduinoReader) ReadData() (vars []*DataHeader, debugs []*DataHeader, events []*DataHeader, err error) {
	if ar.conn == nil {
		return nil, nil, nil, fmt.Errorf("connessione seriale non aperta")
	}

	// Pulisce le liste
	ar.Variables = ar.Variables[:0]
	ar.Debugs = ar.Debugs[:0]
	ar.Events = ar.Events[:0]

	numMessages, err := ar.protocol.ReadCommunicationData()
	if err != nil {
		return nil, nil, nil, fmt.Errorf("errore di lettura dati comunicazione: %w", err)
	}
	if numMessages == 0 {
		return ar.Variables, ar.Debugs, ar.Events, nil // Nessun messaggio da leggere
	}

	for i := 0; i < numMessages; i++ {
		msg, err := ar.protocol.ReadMessage()
		if err != nil {
			return nil, nil, nil, fmt.Errorf("errore durante la lettura del messaggio %d: %w", i+1, err)
		}
		switch msg.MessageType {
		case Var:
			ar.Variables = append(ar.Variables, msg)
		case Debug:
			ar.Debugs = append(ar.Debugs, msg)
		case Event:
			ar.Events = append(ar.Events, msg)
		}
	}

	return ar.Variables, ar.Debugs, ar.Events, nil
}

func (ar *ArduinoReader) WriteData(value int16, id byte) error {
	if ar.conn == nil {
		return fmt.Errorf("connessione seriale non aperta")
	}
	return ar.protocol.WriteData(value, id)
}
