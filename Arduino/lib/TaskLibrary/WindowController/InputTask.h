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
    int isButtonPressed = 0;

public:
    InputTask(
        AnalogInput &potentiometer,
        DigitalInput &manualButton)
        :
          potentiometer(potentiometer),
          manualButton(manualButton)
    {
        ServiceLocator::getSerialManagerInstance().addVariableToSend((byte *)&isButtonPressed, VarType::INT);
    }
    void tick() override
    {
        potentiometer.update();
        manualButton.update();
        isButtonPressed = manualButton.isActive();
    }
    void reset() override
    {
    }
};

#endif