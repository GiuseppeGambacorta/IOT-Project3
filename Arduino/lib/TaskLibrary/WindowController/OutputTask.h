#ifndef __OUTPUTTASK__
#define __OUTPUTTASK__

#include "TaskLibrary.h"
#include "Components.h"
#include "ArduinoStandardLibrary.h"

class OutputTask : public Task
{
private:
    Motor &motor;
    Display &display;
    int motorPosition = 0;

public:
    OutputTask(Motor &motor, Display &display) : motor(motor), display(display)
    {
        ServiceLocator::getSerialManagerInstance().addVariableToSend((byte *)&motorPosition, VarType::INT);
    }
    void tick() override
    {
        motor.update();
        display.update();
        motorPosition = motor.getPosition();
    }
    void reset() override
    {
    }
};

#endif