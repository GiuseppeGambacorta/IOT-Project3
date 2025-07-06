package main

import (
	"encoding/binary"
	"fmt"
	"io"
	"math"
)

type MessageType byte

const (
	Var   MessageType = 0
	Debug MessageType = 1
	Event MessageType = 2
)

type VarType byte

const (
	Byte   VarType = 0
	Int    VarType = 1
	String VarType = 2
	Float  VarType = 3
)

type DataHeader struct {
	MessageType MessageType
	VarType     VarType
	ID          byte
	Size        byte
	// 'any' (o interface{}) permette di contenere tipi diversi (int16, string, float32).
	Data any
}

type Protocol struct {
	conn          io.ReadWriteCloser
	sendBuffer    []byte
	numVarsToSend byte
}

func NewProtocol(conn io.ReadWriteCloser) *Protocol {
	return &Protocol{
		conn:          conn,
		sendBuffer:    make([]byte, 0, 128), // Pre-alloca un po' di spazio
		numVarsToSend: 0,
	}
}

func (p *Protocol) Handshake() error {
	buf := make([]byte, 1)
	for {
		fmt.Println("Aspetto che Arduino si connetta...")

		if _, err := p.conn.Write([]byte{255}); err != nil {
			return fmt.Errorf("errore durante la scrittura per l'handshake: %w", err)
		}

		n, err := p.conn.Read(buf)
		if err != nil {
			if err == io.EOF {
				continue // Timeout, riprova
			}
			return fmt.Errorf("errore durante la lettura per l'handshake: %w", err)
		}
		if n > 0 && buf[0] == 10 {
			break
		}
	}
	fmt.Println("Arduino connesso!")
	return nil
}

func (p *Protocol) readByte() (byte, error) {
	buf := make([]byte, 1)
	_, err := p.conn.Read(buf)
	return buf[0], err
}

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

	numMessages, err := p.readByte()
	if err != nil {
		return 0, err
	}
	return int(numMessages), nil
}

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

func (p *Protocol) AddVariableToSend(id byte, value int16) {
	const size byte = 2 // La dimensione di un int16 è sempre 2 byte

	valueBytes := make([]byte, size)
	binary.LittleEndian.PutUint16(valueBytes, uint16(value))

	// Costruisce il pacchetto per la singola variabile: [ID, Size, Dati...]
	variablePacket := []byte{id, size}
	variablePacket = append(variablePacket, valueBytes...)

	// Aggiunge i dati della variabile al buffer di invio generale
	p.sendBuffer = append(p.sendBuffer, variablePacket...)
	p.numVarsToSend++
}

func (p *Protocol) SendBuffer() error {
	if p.numVarsToSend == 0 {
		return nil
	}

	header := []byte{255, 0, p.numVarsToSend}

	finalPacket := append(header, p.sendBuffer...)

	//fmt.Printf("DEBUG: Inviando pacchetto completo (%d variabili): %v\n", p.numVarsToSend, finalPacket)
	_, err := p.conn.Write(finalPacket)

	p.sendBuffer = p.sendBuffer[:0]
	p.numVarsToSend = 0

	return err
}
