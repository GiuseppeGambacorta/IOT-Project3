#ifndef __WASTE_DISPOSAL_TASK__
#define __WASTE_DISPOSAL_TASK__

#include "Components.h"
#include "TaskLibrary.h"
#include "ArduinoStandardLibrary.h"
#include "Services.h"



enum WindowManagerState : byte {
    AUTOMATIC = 0,
    MANUAL = 1
};

class WindowControllerTask : public Task {

private:

    WindowManagerState state;
    AnalogInput& potentiometer;
    DigitalInput& manualButton;
    Motor& motor;
    Display& display;

    SerialManager &serialManager = ServiceLocator::getSerialManagerInstance();
    RTrig buttonTrigger;

    int *temperature;
    int *actualMode;

    int manualButtonPressed;
    int actualWindowPosition;
    
public:
    WindowControllerTask(

                    AnalogInput& potentiometer,
                    DigitalInput& manualButton,
                    Motor& motor,
                    Display& display);

    void tick() override; 
    void reset() override;

};

#endif