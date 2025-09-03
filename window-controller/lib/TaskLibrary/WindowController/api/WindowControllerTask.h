#ifndef __WASTE_DISPOSAL_TASK__
#define __WASTE_DISPOSAL_TASK__

#include "Components.h"
#include "TaskLibrary.h"
#include "ArduinoStandardLibrary.h"
#include "Services.h"



enum WindowManagerMode : int16_t {
    MANUAL = 0,
    AUTOMATIC = 1
};

enum WindowManagerState : int16_t {
    NORMAL,
    HOT,
    TOO_HOT,
    ALARM
};

enum windowManualCommand : int16_t
{
    NONE = 0,
    UP = 1,
    DOWN = 2
};

class WindowControllerTask : public Task {

private:

    WindowManagerMode oldMode;
    WindowManagerState oldState = NORMAL;
    windowManualCommand oldCommand;
    AnalogInput& potentiometer;
    DigitalInput& manualButton;
    Motor& motor;
    Display& display;

    SerialManager &serialManager = ServiceLocator::getSerialManagerInstance();
    RTrig buttonTrigger;

    int16_t temperature;
    WindowManagerMode actualMode;
    windowManualCommand& windowcommand;
    WindowManagerState actualState;
    int16_t systemWindowPos;

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