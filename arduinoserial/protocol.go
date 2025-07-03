package main

import (
	"encoding/binary"
	"fmt"
	"io"
	"math"
)

// MessageType simula l'enum MessageType di Python.
type MessageType byte

const (
	Var   MessageType = 0
	Debug MessageType = 1
	Event MessageType = 2
)

// VarType simula l'enum VarType di Python.
type VarType byte

const (
	Byte   VarType = 0
	Int    VarType = 1
	String VarType = 2
	Float  VarType = 3
)

// DataHeader contiene i dati di un singolo messaggio, simile alla classe Python.
type DataHeader struct {
	MessageType MessageType
	VarType     VarType
	ID          byte
	Size        byte
	// 'any' (o interface{}) permette di contenere tipi diversi (int16, string, float32).
	Data any
}

// Protocol gestisce la comunicazione a basso livello.
type Protocol struct {
	conn io.ReadWriteCloser
}

// NewProtocol crea una nuova istanza di Protocol.
func NewProtocol(conn io.ReadWriteCloser) *Protocol {
	return &Protocol{conn: conn}
}

// Handshake esegue la stretta di mano con Arduino.
func (p *Protocol) Handshake() error {
	buf := make([]byte, 1)
	for {
		fmt.Println("Aspetto che Arduino si connetta...")
		// Invia 255
		if _, err := p.conn.Write([]byte{255}); err != nil {
			return fmt.Errorf("errore durante la scrittura per l'handshake: %w", err)
		}

		// Legge la risposta
		n, err := p.conn.Read(buf)
		if err != nil {
			if err == io.EOF {
				continue // Timeout, riprova
			}
			return fmt.Errorf("errore durante la lettura per l'handshake: %w", err)
		}
		if n > 0 && buf[0] == 10 {
			break // Successo
		}
	}
	fmt.Println("Arduino connesso!")
	return nil
}

// readByte legge un singolo byte dalla connessione.
func (p *Protocol) readByte() (byte, error) {
	buf := make([]byte, 1)
	_, err := p.conn.Read(buf)
	return buf[0], err
}

// ReadCommunicationData legge l'header della comunicazione e restituisce il numero di messaggi.
func (p *Protocol) ReadCommunicationData() (int, error) {
	// Legge la sequenza di start 255, 0
	b1, err := p.readByte()
	if err != nil {
		return 0, err
	}
	b2, err := p.readByte()
	if err != nil {
		return 0, err
	}

	if b1 != 255 || b2 != 0 {
		return 0, fmt.Errorf("errore di sincronizzazione, ricevuto: %d, %d", b1, b2)
	}

	// Legge il numero di messaggi
	numMessages, err := p.readByte()
	if err != nil {
		return 0, err
	}
	return int(numMessages), nil
}

// ReadMessage legge e decodifica un singolo messaggio.
func (p *Protocol) ReadMessage() (*DataHeader, error) {
	msgType, err := p.readByte()
	if err != nil {
		return nil, err
	}
	varType, err := p.readByte()
	if err != nil {
		return nil, err
	}
	id, err := p.readByte()
	if err != nil {
		return nil, err
	}
	size, err := p.readByte()
	if err != nil {
		return nil, err
	}

	dataBuf := make([]byte, size)
	if _, err := io.ReadFull(p.conn, dataBuf); err != nil {
		return nil, fmt.Errorf("impossibile leggere il payload completo: %w", err)
	}

	header := &DataHeader{
		MessageType: MessageType(msgType),
		VarType:     VarType(varType),
		ID:          id,
		Size:        size,
	}

	switch header.VarType {
	case Int:
		header.Data = int16(binary.LittleEndian.Uint16(dataBuf))
	case String:
		header.Data = string(dataBuf)
	case Float:
		bits := binary.LittleEndian.Uint32(dataBuf)
		header.Data = math.Float32frombits(bits)
	default:
		return nil, fmt.Errorf("tipo di variabile non riconosciuto: %d", header.VarType)
	}

	return header, nil
}

// WriteData scrive un valore intero sulla connessione seriale.
func (p *Protocol) WriteData(value int16, id byte) error {
	// Header di inizio comunicazione
	header := []byte{255, 0}

	// Dati del messaggio
	messageType := byte(Var)
	varType := byte(Int)
	size := byte(2) // int16 occupa 2 byte

	// Converte il valore in byte
	valueBytes := make([]byte, 2)
	binary.BigEndian.PutUint16(valueBytes, uint16(value))

	// Costruisce il pacchetto completo
	packet := append(header, messageType, varType, id, size)
	packet = append(packet, valueBytes...)

	_, err := p.conn.Write(packet)
	return err
}
