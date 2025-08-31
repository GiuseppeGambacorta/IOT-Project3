package mqtt

import (
	"log"
	"strconv"
	"time"

	MQTT "github.com/eclipse/paho.mqtt.golang"
)

func MqttPublishInterval(client MQTT.Client, IntervalUpdatesChan <-chan time.Duration) {
	const configTopic = "esp32/config/interval"
	log.Println("INFO: Publisher MQTT avviato.")
	for interval := range IntervalUpdatesChan {
		intervalPayload := strconv.FormatInt(interval.Milliseconds(), 10)
		token := client.Publish(configTopic, 1, false, intervalPayload)
		token.Wait()
	}
}

func ConfigureClient(broker string, clientID string, onConnectCallbacks ...func(MQTT.Client)) (MQTT.Client, error) {
	opts := MQTT.NewClientOptions()
	opts.AddBroker(broker)
	opts.SetClientID(clientID)
	opts.SetAutoReconnect(true)
	opts.SetConnectRetry(true)
	opts.SetConnectRetryInterval(2 * time.Second)
	opts.SetKeepAlive(30 * time.Second)
	opts.SetPingTimeout(10 * time.Second)
	opts.SetCleanSession(true)

	// Log di connessione/disconnessione
	opts.OnConnect = func(c MQTT.Client) {
		log.Println("MQTT: connesso al broker.")
		for _, callback := range onConnectCallbacks {
			callback(c)
		}
	}
	opts.OnConnectionLost = func(c MQTT.Client, err error) {
		log.Printf("MQTT: connessione persa: %v", err)
	}
	opts.OnReconnecting = func(c MQTT.Client, opts *MQTT.ClientOptions) {
		log.Println("MQTT: tentativo di riconnessione...")
	}

	client := MQTT.NewClient(opts)
	token := client.Connect()
	if token.Wait() && token.Error() != nil {
		return nil, token.Error()
	}
	return client, nil
}
