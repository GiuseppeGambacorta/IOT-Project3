#include <Arduino.h>
#include "../api/Sonar.h"

#define D1 10
const double vs = 331.45 + 0.62*20;

Sonar::Sonar(int triggerPin, int echoPin) {
    this->echoPin = echoPin;
    trigger = new DigitalOutput(triggerPin);
    echo = new DigitalInput(echoPin, 1000);
}

void Sonar::update() {


  
    trigger->turnOn();
    trigger->update();
    delayMicroseconds(10);
    trigger->turnOff();
    trigger->update();

  
    long tUS = pulseInLong(this->echoPin, HIGH, 30000); // 0 if no signal in 30ms
    if (tUS == 0) {
        this->level = -1; // Nessun eco rilevato
        return;
    }

    double t = tUS * 1e-6 / 2; 

 
    this->level = t * vs;
}

float Sonar::readDistance() {
    return this->level;
}

bool Sonar::isThresholdLower() {
    return this->level < D1;
}