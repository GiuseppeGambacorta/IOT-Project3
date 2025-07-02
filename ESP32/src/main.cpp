#include <Arduino.h>
#include <WiFi.h>
#include <PubSubClient.h>
#include "ArduinoStandardLibrary.h" 
#include "TemperatureSensor.h"

// --- CONFIGURAZIONE UTENTE ---
const char* ssid = "TP-LINK_2.4GHz_E0FF27";
const char* password = "gambacorta";
const char* mqtt_server = "192.168.1.191"; 
const int mqtt_port = 1883;

// Topic MQTT
const char* config_topic = "esp32/config/interval";
const char* temp_topic = "esp32/data/temperature";
// -----------------------------

// Oggetti globali
WiFiClient espClient;
PubSubClient client(espClient);
TemperatureSensor tempSensor(A2);
DigitalOutput Redled(17);   // Pin per il LED rosso (problema di connessione)
DigitalOutput Greenled(16); // Pin per il LED verde (connessione OK)

// Dichiarazioni Funzioni e Task
void ledManagerTask(void *pvParameters);
void mqttTask(void *pvParameters);

// Funzione di callback per i messaggi MQTT in arrivo
void callback(char* topic, byte* payload, unsigned int length) {
  Serial.print("Messaggio arrivato sul topic: ");
  Serial.println(topic);
  // Qui andrà la logica per gestire il messaggio
}

// Funzione per la riconnessione a MQTT
void reconnect() {
  while (!client.connected()) {
    Serial.print("Tentativo di connessione MQTT...");
    // "ESP32Client" è l'ID del client. Va benissimo.
    if (client.connect("ESP32Client")) {
      Serial.println("connesso");
      client.subscribe(config_topic);
      Serial.print("Sottoscritto al topic: ");
      Serial.println(config_topic);
    } else {
      Serial.print("fallito, rc=");
      Serial.print(client.state());
      Serial.println(" riprovo tra 5 secondi");
      // Usa vTaskDelay invece di delay in un task
      vTaskDelay(pdMS_TO_TICKS(5000)); 
    }
  }
}

void setup() {
  Serial.begin(115200);
  // Rimuovi while(!Serial); per ESP32-S3 con le build flags
  
  // Connessione al WiFi
  Serial.println();
  Serial.print("Connessione a ");
  Serial.println(ssid);
  WiFi.begin(ssid, password);
  while (WiFi.status() != WL_CONNECTED) {
    delay(500);
    Serial.print(".");
  }
  Serial.println("\nWiFi connesso");
  Serial.print("Indirizzo IP: ");
  Serial.println(WiFi.localIP());

  // Configurazione MQTT
  client.setServer(mqtt_server, mqtt_port);
  client.setCallback(callback);

  // Creazione dei task
  xTaskCreatePinnedToCore(ledManagerTask, "LedManager", 2048, NULL, 2, NULL, 1);
  xTaskCreatePinnedToCore(mqttTask, "MqttTask", 4096, NULL, 1, NULL, 1);
}

// Task per la gestione di connessione e messaggi MQTT
void mqttTask(void *pvParameters) {
  while(true) {
    if (!client.connected()) {
      reconnect();
    }
    client.loop(); // Fondamentale per ricevere messaggi e mantenere la connessione
    vTaskDelay(pdMS_TO_TICKS(100));
  }
}

// Task per la gestione dei LED di stato
void ledManagerTask(void *pvParameters) {
  while (true) {
    // Ora controlla sia WiFi che MQTT per uno stato completo
    if (WiFi.status() != WL_CONNECTED || !client.connected()) {
      Redled.turnOn();
      Greenled.turnOff();
    } else {
      Redled.turnOff();
      Greenled.turnOn();
    }
    Redled.update();
    Greenled.update();
    vTaskDelay(pdMS_TO_TICKS(500));
  }
}

void loop() {
  // Il loop è vuoto, i task fanno tutto il lavoro
  vTaskDelay(portMAX_DELAY);
}