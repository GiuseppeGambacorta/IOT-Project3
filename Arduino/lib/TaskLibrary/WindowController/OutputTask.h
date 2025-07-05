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

public:
    OutputTask(Motor &motor, Display &display) : motor(motor), display(display)
    {
    }
    void tick() override
    {
        motor.update();
        display.update();
    }
    void reset() override
    {
    }
};

#endif