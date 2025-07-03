#pragma once

#include "ArduinoStandardLibrary.h"
#include "Servo.h"

class Motor
{
private:
    unsigned int pin;
    unsigned int offsetPosition;
    int upperLimit;
    int lowerLimit;
    bool initialized = false;
    int commandPosition = 0;
    int lastCommandPosition = 0;
    int lastPosition = 0;
    Servo motor;
    Timer checkPositionTimer = Timer(350);

public:
    Motor(unsigned int pin, unsigned int offsetPosition, int upperLimit, int lowerLimit);
    void init();
    void setPosition(int value);
    int getPosition();
    bool isInPosition();
    bool isAtUpperLimit();
    bool isAtLowerLimit();
    void update();
};

