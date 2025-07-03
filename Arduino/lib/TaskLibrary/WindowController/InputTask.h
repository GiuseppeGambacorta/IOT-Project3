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

public:
    InputTask(
        TemperatureSensor &tempSensor,
        DigitalInput &manualButton)
        :

          tempSensor(tempSensor),
          manualButton(manualButton)
    {
        // ServiceLocator::getSerialManagerInstance().addVariableToSend((byte *)&temp, VarType::INT);
        // ServiceLocator::getSerialManagerInstance().addVariableToSend((byte *)&level, VarType::FLOAT);
    }
    void tick() override
    {
        tempSensor.update();
        manualButton.update();
    }
    void reset() override
    {
    }
};

#endif