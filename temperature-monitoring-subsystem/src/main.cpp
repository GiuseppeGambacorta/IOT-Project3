#include <Arduino.h>
#include <WiFi.h>
#include <PubSubClient.h>
#include "ArduinoStandardLibrary.h" 
#include "TemperatureSensor.h"

const char* ssid = "TP-LINK_2.4GHz_E0FF27";
const char* password = "gambacorta";
const char* mqtt_server = "192.168.1.103"; 
const int mqtt_port = 1883;

// Topic MQTT
const char* config_topic = "esp32/config/interval";
const char* temp_topic = "esp32/data/temperature";
// -----------------------------

WiFiClient espClient;
PubSubClient client(espClient);
TemperatureSensor tempSensor(18,4095 , 100, 55);
DigitalOutput Redled(17);   
DigitalOutput Greenled(16); 

// Variabili condivise e Mutex per la protezione
volatile float currentTemperature = 0.0;
volatile long publishInterval = 0; // Intervallo in ms. Se 0, non si pubblica.
SemaphoreHandle_t frequencyMutex;
SemaphoreHandle_t temperatureMutex;

void ledManagerTask(void *pvParameters);
void readTempTask(void *pvParameters);
void publishMqttTask(void *pvParameters);

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
      xSemaphoreTake(frequencyMutex, portMAX_DELAY);
      publishInterval = newInterval;
      xSemaphoreGive(frequencyMutex);
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
      Serial.println(" riprovo tra 1 secondo");
      vTaskDelay(pdMS_TO_TICKS(1000)); 
    }
  }
}

void readTempTask(void *pvParameters) {
  while (true) {
    long localInterval;
    xSemaphoreTake(frequencyMutex, portMAX_DELAY);
    localInterval = publishInterval;
    xSemaphoreGive(frequencyMutex);

    if (localInterval > 0) {
      tempSensor.update();
      float temp = tempSensor.readTemperature();
      
      xSemaphoreTake(temperatureMutex, portMAX_DELAY);
      currentTemperature = temp;
      xSemaphoreGive(temperatureMutex);
      
      vTaskDelay(pdMS_TO_TICKS(localInterval));
    } else {
      vTaskDelay(pdMS_TO_TICKS(1000));
    }
  }
}

void publishMqttTask(void *pvParameters) {
  while (true) {
    if (!client.connected()) {
      reconnect();
    }
    client.loop(); // Gestisce i messaggi in arrivo (es. la callback)

    long localInterval;
    xSemaphoreTake(frequencyMutex, portMAX_DELAY);
    localInterval = publishInterval;
    xSemaphoreGive(frequencyMutex);

    if (localInterval > 0) {
      float tempToPublish;
      xSemaphoreTake(temperatureMutex, portMAX_DELAY);
      tempToPublish = currentTemperature;
      xSemaphoreGive(temperatureMutex);
      char tempString[8];
      dtostrf(tempToPublish, 4, 2, tempString);
      
      Serial.print("Invio temperatura: ");
      Serial.print(tempString);
      Serial.print(" al topic: ");
      Serial.println(temp_topic);

      client.publish(temp_topic, tempString);
      vTaskDelay(pdMS_TO_TICKS(localInterval));
    } else {
      vTaskDelay(pdMS_TO_TICKS(200));
    }
  }
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
    Serial.println("led task is cycling");
    Redled.update();
    Greenled.update();
    vTaskDelay(pdMS_TO_TICKS(500));
  }
}

void setup() {
  Serial.begin(115200);
  
  frequencyMutex = xSemaphoreCreateMutex(); 
  temperatureMutex = xSemaphoreCreateMutex(); 
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

  xTaskCreatePinnedToCore(readTempTask, "ReadTemp", 2048, NULL, 1, NULL, 0);
  xTaskCreatePinnedToCore(ledManagerTask, "LedManager", 2048, NULL, 2, NULL, 1);
  xTaskCreatePinnedToCore(publishMqttTask, "PublishMQTT", 4096, NULL, 1, NULL, 1);
}



void loop() {
  // Il loop è vuoto, i task fanno tutto il lavoro
  vTaskDelay(portMAX_DELAY);
}