#ifndef __WASTE_DISPOSAL_TASK__
#define __WASTE_DISPOSAL_TASK__

#include "Components/Pir/Api/Pir.h"
#include "Components/Sonar/Api/Sonar.h"
#include "Components/Temperaturesensor/Api/TemperatureSensor.h"

#include "subTask/StdExecTask.h"
#include "subTask/AlarmLevelTask.h"
#include "subTask/AlarmTempTask.h"


enum ResetMessage : int {
    NO_MESSAGE = 0,
    MESSAGE_FROM_GUI = 1,
    MESSAGE_SEEN = 2,
};

enum WasteDisposalState {
    STD_EXEC,
    LVL_ALLARM,
    TEMP_ALLARM,

};

class WasteDisposalTask : public Task {

private:
    StdExecTask& stdExecTask;
    AlarmLevelTask& alarmLevelTask;
    AlarmTempTask& alarmTempTask;

    WasteDisposalState state;
    WasteDisposalState oldState;
    Timer tempTimer;
    Timer emptyTimer;

    Sonar& levelDetector;
    TemperatureSensor& tempSensor;

    int *empty;
    int *fire;
    
public:
    WasteDisposalTask(
                    StdExecTask& stdExecTask,
                    AlarmLevelTask& alarmLevelTask,
                    AlarmTempTask& alarmTempTask,
                    Sonar& levelDetector,
                    TemperatureSensor& tempSensor);

    void tick() override; 
    void reset() override;

    const char* wasteDisposalStateToString(WasteDisposalState state) {
    switch (state) {
        case STD_EXEC: return "STD_EXEC";
        case LVL_ALLARM: return "LVL_ALLARM";
        case TEMP_ALLARM: return "TEMP_ALLARM";
        default: return "UNKNOWN";
    }
}
};

#endif