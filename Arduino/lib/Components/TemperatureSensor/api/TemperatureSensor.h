#pragma once
#include <ArduinoStandardLibrary.h>

class TemperatureSensor{
    public:
        TemperatureSensor(unsigned int pin, unsigned int maxRange,unsigned int mapRange, int offset);
        void update();
        int readTemperature();
        bool isThresholdExceeded();
    private:
        int temperature;
        AnalogInput* sensor;
};
