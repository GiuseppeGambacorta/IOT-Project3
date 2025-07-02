#include <Arduino.h>
#include <WiFi.h>
#include <PubSubClient.h>
#include "TemperatureSensor.h"
#include "ArduinoStandardLibrary.h" 

// --- CONFIGURAZIONE UTENTE ---
const char* ssid = "TP-LINK_5GHz_E0FF29";
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
DigitalOutput Redled(1);   // Pin per il LED rosso (problema di connessione)
DigitalOutput Greenled(2); // Pin per il LED verde (connessione OK)


// Variabili condivise tra i task
volatile float currentTemperature = 0.0;
volatile long publishInterval = 0; // Intervallo in ms, 0 = non inviare

// Mutex per l'accesso sicuro alle variabili condivise
SemaphoreHandle_t dataMutex;

// Dichiarazioni delle funzioni dei task
void readTempTask(void *pvParameters);
void publishMqttTask(void *pvParameters);
void ledManagerTask(void *pvParameters); // Nuovo task per i LED

// Funzione di callback per i messaggi MQTT in arrivo
void callback(char* topic, byte* payload, unsigned int length) {
  Serial.print("Messaggio arrivato sul topic: ");
  Serial.println(topic);

  char message[length + 1];
  memcpy(message, payload, length);
  message[length] = '\0';

  if (strcmp(topic, config_topic) == 0) {
    long newInterval = atol(message);
    if (newInterval > 0) {
      xSemaphoreTake(dataMutex, portMAX_DELAY);
      publishInterval = newInterval;
      xSemaphoreGive(dataMutex);
      Serial.print("Nuovo intervallo di pubblicazione impostato a: ");
      Serial.print(publishInterval);
      Serial.println(" ms");
    }
  }
}

void reconnect() {
  while (!client.connected()) {
    Serial.print("Tentativo di connessione MQTT...");
    if (client.connect("ESP32Client")) {
      Serial.println("connesso");
      client.subscribe(config_topic);
      Serial.print("Sottoscritto al topic: ");
      Serial.println(config_topic);
    } else {
      Serial.print("fallito, rc=");
      Serial.print(client.state());
      Serial.println(" riprovo tra 5 secondi");
      delay(5000);
    }
  }
}

void setup() {
  Serial.begin(9600);
  delay(10);
  Serial.print("Prova");
  // Inizializzazione LED: rosso acceso, verde spento
  Redled.turnOn();
  Greenled.turnOff();
  Redled.update();
  Greenled.update();

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

  // Crea il mutex
  dataMutex = xSemaphoreCreateMutex();

  // Creazione dei task su core specifici
  //xTaskCreatePinnedToCore(readTempTask, "ReadTemp", 2048, NULL, 1, NULL, 0);
  //xTaskCreatePinnedToCore(publishMqttTask, "PublishMQTT", 4096, NULL, 1, NULL, 1);
  //xTaskCreatePinnedToCore(ledManagerTask, "LedManager", 1024, NULL, 2, NULL, 0); // Aggiunto task per i LED
}

// Task per la gestione dei LED di stato (Core 1)
void ledManagerTask(void *pvParameters) {
  
  while (true) {
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

// Task per la lettura della temperatura (Core 0)
void readTempTask(void *pvParameters) {
  while (true) {
    long localInterval;
    xSemaphoreTake(dataMutex, portMAX_DELAY);
    localInterval = publishInterval;
    xSemaphoreGive(dataMutex);

    if (localInterval > 0) {
      tempSensor.update();
      float temp = tempSensor.readTemperature();
      xSemaphoreTake(dataMutex, portMAX_DELAY);
      currentTemperature = temp;
      xSemaphoreGive(dataMutex);
      vTaskDelay(pdMS_TO_TICKS(localInterval));
    } else {
      vTaskDelay(pdMS_TO_TICKS(1000));
    }
  }
}

// Task per l'invio MQTT (Core 1)
void publishMqttTask(void *pvParameters) {
  while (true) {
    long localInterval;
    float tempToPublish;

    xSemaphoreTake(dataMutex, portMAX_DELAY);
    localInterval = publishInterval;
    xSemaphoreGive(dataMutex);

    if (localInterval > 0) {
      if (!client.connected()) {
        reconnect();
      }
      client.loop();

      xSemaphoreTake(dataMutex, portMAX_DELAY);
      tempToPublish = currentTemperature;
      xSemaphoreGive(dataMutex);

      char tempString[8];
      dtostrf(tempToPublish, 4, 2, tempString);
      
      Serial.print("Invio temperatura: ");
      Serial.print(tempString);
      Serial.print(" al topic: ");
      Serial.println(temp_topic);

      client.publish(temp_topic, tempString);
      vTaskDelay(pdMS_TO_TICKS(localInterval));
    } else {
      vTaskDelay(pdMS_TO_TICKS(1000));
    }
  }
}

void loop() {
  // Il loop è vuoto, i task fanno tutto il lavoro
  vTaskDelay(portMAX_DELAY);
}