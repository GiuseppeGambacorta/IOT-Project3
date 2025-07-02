#include "TemperatureSensor.h"
#include <Arduino.h>

#define MAXTEMP 100
#define TEMPOFFSET 55

TemperatureSensor::TemperatureSensor(int pin) {
    this->temperature = 0;
    this->sensor = new AnalogInput(pin, 206);
}

void TemperatureSensor::update() {
    sensor->update();
    int analogValue = sensor->getValue() - TEMPOFFSET;
    this->temperature = analogValue;
}

int TemperatureSensor::readTemperature() {
    return this->temperature;
}

bool TemperatureSensor::isThresholdExceeded() {
    return this->temperature > MAXTEMP;
}