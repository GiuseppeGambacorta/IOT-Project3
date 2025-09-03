#include "TemperatureSensor.h"
#include <Arduino.h>

#define MAXTEMP 100
#define TEMPOFFSET 55

TemperatureSensor::TemperatureSensor(unsigned int pin, unsigned int maxRange,unsigned int mapRange, int offset) {
    this->temperature = 0;
    this->sensor = new AnalogInput(pin, maxRange, mapRange, offset);
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