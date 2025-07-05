#ifndef __INPUTTASK__
#define __INPUTTASK__

#include "TaskLibrary.h"
#include "Components.h"
#include "ArduinoStandardLibrary.h"

class InputTask : public Task
{
private:
    AnalogInput &potentiometer;
    DigitalInput &manualButton;

public:
    InputTask(
        AnalogInput &potentiometer,
        DigitalInput &manualButton)
        :
          potentiometer(potentiometer),
          manualButton(manualButton)
    {
        
    }
    void tick() override
    {
        potentiometer.update();
        manualButton.update();
    }
    void reset() override
    {
    }
};

#endif