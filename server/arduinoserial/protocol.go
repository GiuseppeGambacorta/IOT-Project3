package arduinoserial

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

type Message struct {
	MessageType MessageType
	VarType     VarType
	ID          byte
	Size        byte
	Data        any
}

type Protocol struct {
	conn          io.ReadWriteCloser
	dataToSend    []byte
	numVarsToSend byte
}

func NewProtocol(conn io.ReadWriteCloser) *Protocol {
	return &Protocol{
		conn:          conn,
		dataToSend:    make([]byte, 0, 128),
		numVarsToSend: 0,
	}
}

// send a 255 byte handshake to Arduino and wait for a response, the responde should be a byte 10
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

// the first two bytes should be 255 and 0, the third byte is the number of messages
func (p *Protocol) ReadCommunicationData() (int, error) {

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

func (p *Protocol) ReadMessage() (*Message, error) {
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

	header := &Message{
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

func (p *Protocol) AddVariableToSend(id byte, value []byte) {

	size := byte(len(value))
	variablePacket := []byte{id, size}
	variablePacket = append(variablePacket, value...)

	p.dataToSend = append(p.dataToSend, variablePacket...)
	p.numVarsToSend++
}

func (p *Protocol) SendBuffer() error {
	if p.numVarsToSend == 0 {
		return nil
	}

	header := []byte{255, 0, p.numVarsToSend}

	finalPacket := append(header, p.dataToSend...)

	fmt.Printf("DEBUG: Inviando pacchetto completo (%d variabili): %v\n", p.numVarsToSend, finalPacket)
	_, err := p.conn.Write(finalPacket)

	p.dataToSend = p.dataToSend[:0]
	p.numVarsToSend = 0

	return err
}
