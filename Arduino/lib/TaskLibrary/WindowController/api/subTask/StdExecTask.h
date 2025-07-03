#ifndef __STDEXECTASK__
#define __STDEXECTASK__

#include "../../../Task.h"

#include "ArduinoStandardLibrary.h"
#include "Components/Door/Api/Door.h"
#include "Components/Display/Api/Display.h"
#include "Components/Pir/Api/Pir.h"


#define TOpen 5000
#define TSleep 10000
#define TClose 10000

enum StdExecState{
    READY,
    OPEN,
    SLEEP
};

class StdExecTask : public Task {
    
    private:
        StdExecState state;

        Door& door;
        Display& display;
        DigitalInput& openButton;
        DigitalInput& closeButton;
        DigitalOutput& ledGreen;
        DigitalOutput& ledRed;
        Pir& userDetector;

        bool userStatus;
        Timer openTimer;
        Timer userTimer;
        Timer closeTimer;

        void homingReady();
        void homingOpen();
        void homingSleep();

        void execReady();
        void execOpen();
        void execSleep();

    public:
        StdExecTask(Door& door,
                    Display& display,
                    DigitalInput& openButton,
                    DigitalInput& closeButton,
                    DigitalOutput& ledGreen,
                    DigitalOutput& ledRed,
                    Pir& userDetector);
        void tick() override;
        void reset() override;
};

#endif