package arduinoserial

import (
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"server/system"
	"time"

	"go.bug.st/serial"
)

// --- Tipi per la comunicazione con Arduino ---

type DataFromArduino struct {
	WindowPosition system.Degree
	buttonPressed  bool
}

type DataToArduino struct {
	Temperature          int
	OperativeMode        int // 0 per AUTOMATIC, 1 per MANUAL
	WindowAction         int // 0: None, 1: Open, 2: Close
	SystemState          int
	SystemWindowPosition system.Degree
}

type Arduino struct {
	portName string
	baudrate int
	timeout  time.Duration
	protocol *Protocol

	Variables []Message
	Debugs    []Message
	Events    []Message
}

func ManageArduino(dataFromArduino chan DataFromArduino, dataToArduino <-chan DataToArduino) {
	var arduino *Arduino
	var err error
	for {
		arduino, err = createArduino(9600, 2*time.Second)
		if err != nil {
			log.Println("errore %w, riprovo ricerca", err)
		} else {
			break
		}
	}

	// Goroutine per la scrittura
	go func() {
		byteToSend := make([]byte, 2)

		for cmd := range dataToArduino {
			binary.LittleEndian.PutUint16(byteToSend, uint16(cmd.Temperature))
			arduino.AddDataToSend(0, byteToSend)
			binary.LittleEndian.PutUint16(byteToSend, uint16(cmd.OperativeMode))
			arduino.AddDataToSend(1, byteToSend)
			binary.LittleEndian.PutUint16(byteToSend, uint16(cmd.WindowAction))
			arduino.AddDataToSend(2, byteToSend)
			binary.LittleEndian.PutUint16(byteToSend, uint16(cmd.SystemState))
			arduino.AddDataToSend(3, byteToSend)
			binary.LittleEndian.PutUint16(byteToSend, uint16(cmd.SystemWindowPosition))
			arduino.AddDataToSend(4, byteToSend)
			if err := arduino.WriteData(); err != nil {
				log.Printf("ERRORE: Impossibile inviare dati ad Arduino: %v", err)
			} else {
				log.Printf("Dati inviati correttamente ad arduino")
			}
		}
	}()

	// Loop principale per la lettura continua da Arduino
	wasButtonPressed := false
	for {
		vars, _, _, err := arduino.ReadData()
		if err != nil {
			log.Println(err)
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

		newData := DataFromArduino{WindowPosition: system.Degree(windowPos), buttonPressed: bool(buttonState == 1)}

		// Rileva il fronte di salita del pulsante per inviare un solo comando
		if newData.buttonPressed && !wasButtonPressed {
			log.Println("INFO: Pressione pulsante rilevata, invio comando ToggleMode.")
		}
		wasButtonPressed = newData.buttonPressed
		log.Println("dati arrivati")

		select {
		case dataFromArduino <- newData:
			log.Println(newData.WindowPosition)

		default:
			// Il canale è pieno scarto un valore e ne inserisco un altro. Runtime gestisce raceCondition in lettura sul canale, nessun problema di deadlock facendo cosi
			<-dataFromArduino
			dataFromArduino <- newData
			log.Println("WARN: Buffer dati da Arduino pieno. Scartato il valore più vecchio per inserire il più recente.")
		}
	}

}

func createArduino(baudRate int, readTimeout time.Duration) (*Arduino, error) {
	for {
		log.Println("Searching for arduino port")
		arduinoConn, portName, err := findArduinoPort(baudRate, readTimeout)
		if err != nil {
			return nil, err
		}
		if arduinoConn != nil {
			log.Println("Found arduino port: " + portName)
			arduino := &Arduino{
				portName: portName,
				baudrate: baudRate,
				timeout:  readTimeout,
				protocol: NewProtocol(arduinoConn),
			}
			log.Println("INFO: Connesso ad Arduino.")
			return arduino, nil
		}
		time.Sleep(1 * time.Second)
	}
}

func findArduinoPort(baudRate int, readTimeout time.Duration) (io.ReadWriteCloser, string, error) {
	ports, err := getSerialPorts()
	if err != nil {
		return nil, "", fmt.Errorf("errore nella ricerca delle porte: %w", err)
	}

	for _, port := range ports {
		conn, err := Handshake(port, baudRate, readTimeout)
		if err == nil && conn != nil {
			return conn, port, nil
		}
	}
	return nil, "", err

}

func getSerialPorts() ([]string, error) {
	ports, err := serial.GetPortsList()
	if err != nil {
		return nil, err
	}
	if len(ports) == 0 {
		return nil, fmt.Errorf("nessuna porta seriale trovata")
	}
	return ports, nil
}

func (ar *Arduino) Disconnect() {
	if ar.protocol != nil {
		ar.protocol.conn.Close()
		fmt.Println("Connessione chiusa.")
		ar.protocol = nil
	}
}

func (ar *Arduino) ReadData() (vars []Message, debugs []Message, events []Message, err error) {
	if ar.protocol == nil {
		return nil, nil, nil, fmt.Errorf("connessione seriale non aperta")
	}

	//clean slices
	ar.Variables = ar.Variables[:0]
	ar.Debugs = ar.Debugs[:0]
	ar.Events = ar.Events[:0]

	numMessages, err := ar.protocol.ReadCommunicationData()
	if err != nil {
		return nil, nil, nil, fmt.Errorf("errore di lettura dati comunicazione: %w", err)
	}
	if numMessages == 0 {
		return nil, nil, nil, fmt.Errorf("errore di lettura dati comunicazione: 0 messaggi in arrivo")
	}

	for i := 0; i < numMessages; i++ {
		msg, err := ar.protocol.ReadMessage()
		if err != nil {
			return nil, nil, nil, fmt.Errorf("errore durante la lettura del messaggio %d: %w", i+1, err)
		}
		switch msg.MessageType {
		case Var:
			ar.Variables = append(ar.Variables, *msg)
		case Debug:
			ar.Debugs = append(ar.Debugs, *msg)
		case Event:
			ar.Events = append(ar.Events, *msg)
		}
	}

	return ar.Variables, ar.Debugs, ar.Events, nil
}

func (ar *Arduino) AddDataToSend(id byte, value []byte) {
	if ar.protocol == nil {
		fmt.Println("Protocollo non inizializzato, impossibile aggiungere dati.")
		return
	}
	ar.protocol.AddVariableToSend(id, value)
}

func (ar *Arduino) WriteData() error {
	if ar.protocol == nil {
		return fmt.Errorf("protocollo non inizializzato, impossibile aggiungere dati")
	}
	return ar.protocol.SendBuffer()
}
