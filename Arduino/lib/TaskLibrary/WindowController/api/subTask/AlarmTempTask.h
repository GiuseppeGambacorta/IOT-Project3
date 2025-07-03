#ifndef __ALARMTEMPTASK__
#define __ALARMTEMPTASK__

#include "../../../Task.h"
#include "ArduinoStandardLibrary.h"
#include "Components/TemperatureSensor/Api/TemperatureSensor.h"
#include "Components/Door/Api/Door.h"
#include "Components/Display/Api/Display.h"

class AlarmTempTask : public Task {

    private:
        DigitalOutput& ledGreen;
        DigitalOutput& ledRed;
        Display& display;
        Door& door;
        TemperatureSensor& tempSensor;
        enum State {IDLE, ALARM, RESET} state;
        Timer* timer;
        int *fire;

        void handleIdleState();
        void handleAlarmState();
        void handleResetState();
    public:
        AlarmTempTask(DigitalOutput& ledGreen,
                                DigitalOutput& ledRed,
                                Display& display,
                                Door& door,
                                TemperatureSensor& tempSensor);
        void tick() override;
        void reset() override;
};

#endif