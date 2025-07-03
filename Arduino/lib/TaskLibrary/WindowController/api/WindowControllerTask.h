#ifndef __WASTE_DISPOSAL_TASK__
#define __WASTE_DISPOSAL_TASK__

#include "Components.h"
#include "TaskLibrary.h"
#include "ArduinoStandardLibrary.h"
#include "Services.h"



enum WindowManagerState {
    AUTOMATIC,
    MANUAL
};

class WindowControllerTask : public Task {

private:

    WindowManagerState state;
    AnalogInput& potentiometer;
    DigitalInput& manualButton;
    Motor& motor;
    Display& display;

    int *temperature;
    int *actualMode;
    
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