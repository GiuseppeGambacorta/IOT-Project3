#ifndef __OUTPUTTASK__
#define __OUTPUTTASK__

#include "../../../Task.h"

#include "Components/Door/Api/Door.h"
#include "Components/Display/Api/Display.h"
#include "ArduinoStandardLibrary.h"

class OutputTask : public Task {
private:
    Door& door;
    Display& display;
    DigitalOutput& ledGreen;
    DigitalOutput& ledRed;
    SerialManager& serialManager = ServiceLocator::getSerialManagerInstance();

public:
    OutputTask( 
                Door& door,
                Display& display,
              DigitalOutput& ledGreen, 
              DigitalOutput& ledRed);
    void tick() override;
    void reset() override;
};

#endif