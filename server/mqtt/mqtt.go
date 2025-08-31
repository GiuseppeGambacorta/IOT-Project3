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

func ConfigureClient(broker string) MQTT.Client {
	opts := MQTT.NewClientOptions().AddBroker(broker).SetClientID("go-server-logic")
	client := MQTT.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		log.Printf("ERRORE: Impossibile connettersi al broker MQTT: %v", token.Error())
		return nil
	}
	log.Println("INFO: Connesso al broker MQTT.")
	return client
}

func SubscribeToTopic(client MQTT.Client, messageHandler MQTT.MessageHandler, topic string) error {

	if token := client.Subscribe(topic, 1, messageHandler); token.Wait() && token.Error() != nil {
		log.Fatalf("ERRORE: Impossibile sottoscriversi al topic %s: %v", topic, token.Error())
		return token.Error()
	}
	log.Printf("INFO: Sottoscritto con successo al topic: %s\n", topic)
	return nil
}
