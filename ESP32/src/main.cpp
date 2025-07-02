#include <Arduino.h>
#include <WiFi.h>
#include "ArduinoStandardLibrary.h" 

// --- CONFIGURAZIONE UTENTE ---
const char* ssid = "TP-LINK_2.4GHz_E0FF27";
const char* password = "gambacorta";
// -----------------------------

// Oggetti globali
DigitalOutput Redled(17);   // Pin per il LED rosso (problema di connessione)
DigitalOutput Greenled(16); // Pin per il LED verde (connessione OK)

// Dichiarazione del task per i LED
void ledManagerTask(void *pvParameters);

void setup() {
  Serial.begin(115200);
  while(!Serial);
  
  // Inizializzazione LED: rosso acceso, verde spento durante il tentativo
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

  // Creazione del task per la gestione dei LED
  // Aumentiamo lo stack a 2048 per sicurezza ed eseguiamolo sul core 1
  xTaskCreatePinnedToCore(ledManagerTask, "LedManager", 2048, NULL, 1, NULL, 1);
}

// Task per la gestione dei LED di stato
void ledManagerTask(void *pvParameters) {
  while (true) {
    // Controlla solo lo stato del WiFi
    if (WiFi.status() != WL_CONNECTED) {
      Redled.turnOn();
      Greenled.turnOff();
    } else {
      Redled.turnOff();
      Greenled.turnOn();
    }
    Redled.update();
    Greenled.update();
    Serial.println(WiFi.status());
    vTaskDelay(pdMS_TO_TICKS(500)); // Attendi 500ms
  }
}

void loop() {
  // Il loop è vuoto, il task fa tutto il lavoro
  vTaskDelay(portMAX_DELAY);
}