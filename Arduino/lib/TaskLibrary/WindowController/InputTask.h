#ifndef __INPUTTASK__
#define __INPUTTASK__

#include "TaskLibrary.h"
#include "Components.h"
#include "ArduinoStandardLibrary.h"

class InputTask : public Task
{
private:
    TemperatureSensor &tempSensor;
    DigitalInput &manualButton;
    int isButtonPressed = 0;

public:
    InputTask(
        TemperatureSensor &tempSensor,
        DigitalInput &manualButton)
        :

          tempSensor(tempSensor),
          manualButton(manualButton)
    {
        ServiceLocator::getSerialManagerInstance().addVariableToSend((byte *)&isButtonPressed, VarType::INT);
    }
    void tick() override
    {
        tempSensor.update();
        manualButton.update();
        isButtonPressed = manualButton.isActive();
    }
    void reset() override
    {
    }
};

#endif