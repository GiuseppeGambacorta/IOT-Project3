#pragma once

#include <ArduinoStandardLibrary.h>

class TemperatureSensor{
    public:
        TemperatureSensor(int pin);
        void update();
        int readTemperature();
        bool isThresholdExceeded();
    private:
        int temperature;
        AnalogInput* sensor;
};
