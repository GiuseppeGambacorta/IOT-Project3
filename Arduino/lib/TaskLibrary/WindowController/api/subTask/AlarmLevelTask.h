#ifndef __ALARMLEVELTASK__
#define __ALARMLEVELTASK__

#include "../../../Task.h"

#include "ArduinoStandardLibrary.h"
#include "Components/Door/Api/Door.h"
#include "Components/Display/Api/Display.h"
#include <Components/Sonar/api/Sonar.h>

class AlarmLevelTask : public Task {

private:
    Door& door;
    Display& display;
    DigitalOutput& ledGreen;
    DigitalOutput& ledRed;
    Sonar& levelDetector;
    enum State {IDLE, ALARM, EMPTY, RESET} state;
    Timer timer;
    int *empty;

    void handleIdleState();
    void handleAlarmState();
    void handleEmptyState();
    void handleResetState();

public:
    AlarmLevelTask(Door& door,
                    Display& display,
                    DigitalOutput& ledGreen,
                    DigitalOutput& ledRed,
                    Sonar& levelDetector);
    void tick() override;
    void reset() override;
};

#endif