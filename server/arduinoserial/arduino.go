package arduinoserial

import (
	"encoding/binary"
	"fmt"
	"log"
	"server/system"
	"time"

	"go.bug.st/serial"
)

// --- Tipi per la comunicazione con Arduino ---

type DataFromArduino struct {
	WindowPosition int
}

type DataToArduino struct {
	Temperature   int16
	OperativeMode int16 // 0 per AUTOMATIC, 1 per MANUAL
	WindowAction  int16 // 0: None, 1: Open, 2: Close
	SystemState   int16
}

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

	ar.Variables = ar.Variables[:0]
	ar.Debugs = ar.Debugs[:0]
	ar.Events = ar.Events[:0]

	numMessages, err := ar.protocol.ReadCommunicationData()
	if err != nil {
		return nil, nil, nil, fmt.Errorf("errore di lettura dati comunicazione: %w", err)
	}
	if numMessages == 0 {
		return ar.Variables, ar.Debugs, ar.Events, nil
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

func (ar *ArduinoReader) AddDataToSend(id byte, value []byte) {
	if ar.protocol == nil {
		fmt.Println("Protocollo non inizializzato, impossibile aggiungere dati.")
		return
	}
	ar.protocol.AddVariableToSend(id, value)
}

func (ar *ArduinoReader) WriteData() error {
	if ar.conn == nil {
		return fmt.Errorf("connessione seriale non aperta")
	}
	return ar.protocol.SendBuffer()
}

func ManageArduino(requestChan chan<- system.RequestType, dataFromArduino chan<- DataFromArduino, dataToArduino <-chan DataToArduino) {
	arduino := NewArduinoReader(9600, 5*time.Second)
	if err := arduino.Connect(); err != nil {
		log.Printf("ERRORE: Impossibile connettersi ad Arduino: %v. Riprovo...", err)
		time.Sleep(5 * time.Second)
		ManageArduino(requestChan, dataFromArduino, dataToArduino)
		return
	}
	defer arduino.Disconnect()
	log.Println("INFO: Connesso ad Arduino.")

	var wasButtonPressed bool = false

	// Goroutine per la scrittura: si attiva solo quando riceve un comando.
	go func() {
		byteToSend := make([]byte, 2)

		for cmd := range dataToArduino {
			binary.LittleEndian.PutUint16(byteToSend, uint16(cmd.Temperature))
			arduino.AddDataToSend(0, byteToSend)
			binary.LittleEndian.PutUint16(byteToSend, uint16(cmd.OperativeMode))
			arduino.AddDataToSend(1, byteToSend)
			binary.LittleEndian.PutUint16(byteToSend, uint16(cmd.WindowAction))
			arduino.AddDataToSend(2, byteToSend)

			if err := arduino.WriteData(); err != nil {
				log.Printf("ERRORE: Impossibile inviare dati ad Arduino: %v", err)
			}
		}
	}()

	// Loop principale per la lettura continua da Arduino
	for {
		vars, _, _, err := arduino.ReadData()
		if err != nil {
			//timeout, is not critical
			continue
		}

		if len(vars) < 2 {
			log.Println("WARN: Ricevuto pacchetto incompleto da Arduino.")
			continue
		}

		buttonState, ok1 := vars[0].Data.(int16)
		windowPos, ok2 := vars[1].Data.(int16)
		if !ok1 || !ok2 {
			log.Println("ERRORE: Dati da Arduino non validi o tipo inatteso.")
			continue
		}

		isButtonPressed := (buttonState == 1)
		// Rileva il fronte di salita del pulsante per inviare un solo comando
		if isButtonPressed && !wasButtonPressed {
			log.Println("INFO: Pressione pulsante rilevata, invio comando ToggleMode.")
			requestChan <- system.ToggleMode
		}
		wasButtonPressed = isButtonPressed

		dataFromArduino <- DataFromArduino{WindowPosition: int(windowPos)}
	}
}
