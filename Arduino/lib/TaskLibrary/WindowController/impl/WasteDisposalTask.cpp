#include "../api/WasteDisposalTask.h"


#define MAXTEMPTIME 10000
#define TSleep 10000
#define TEmpty 3000
double maxLevel = 0.3;
int maxTemp = 100;

WasteDisposalTask::WasteDisposalTask(StdExecTask& stdExecTask,
                                     AlarmLevelTask& alarmLevelTask,
                                     AlarmTempTask& alarmTempTask,
                                     Sonar& levelDetector,
                                     TemperatureSensor& tempSensor)
            :stdExecTask(stdExecTask),
            alarmLevelTask(alarmLevelTask),
            alarmTempTask(alarmTempTask),
            state(WasteDisposalState::STD_EXEC),
            tempTimer(Timer(MAXTEMPTIME)),
            emptyTimer(Timer(TEmpty)),
            levelDetector(levelDetector),
            tempSensor(tempSensor){
                empty = ServiceLocator::getSerialManagerInstance().getvar(1);
                fire = ServiceLocator::getSerialManagerInstance().getvar(0);
            }

void WasteDisposalTask::tick() {
    double level = levelDetector.readDistance();
    int temp = tempSensor.readTemperature();
    tempTimer.active(temp >= maxTemp);

    switch (state){
    case STD_EXEC:
               

        if (level <= maxLevel) {
            state = WasteDisposalState::LVL_ALLARM;
        }

        if (tempTimer.isTimeElapsed()) {
            state = WasteDisposalState::TEMP_ALLARM;
        }
        break;
    case LVL_ALLARM:
        if (*empty ==  ResetMessage::MESSAGE_SEEN) {
            state = WasteDisposalState::STD_EXEC;
            *empty =  ResetMessage::NO_MESSAGE;
        }
        
        if (tempTimer.isTimeElapsed()) {
            state = WasteDisposalState::TEMP_ALLARM;
        }
        break;
    case TEMP_ALLARM:

        if ( *fire ==  ResetMessage::MESSAGE_SEEN) {
            state = WasteDisposalState::STD_EXEC;
            *fire = ResetMessage::NO_MESSAGE;;
        }
   
    }

    switch (state){
        case STD_EXEC:
            stdExecTask.setActive(true);
            alarmLevelTask.setActive(false);
            alarmTempTask.setActive(false);
            break;
        case LVL_ALLARM:
            stdExecTask.setActive(false);
            alarmLevelTask.setActive(true);
            alarmTempTask.setActive(false);
            break;
        case TEMP_ALLARM:
            stdExecTask.setActive(false);
            alarmLevelTask.setActive(false);
            alarmTempTask.setActive(true);
            break;
    }

    if (state != oldState){
         ServiceLocator::getSerialManagerInstance().addEventMessage(wasteDisposalStateToString(state));
    }
    oldState = state;
}

void WasteDisposalTask::reset() {
    state = WasteDisposalState::STD_EXEC;
    tempTimer.reset();
    stdExecTask.setActive(true);
    alarmLevelTask.setActive(false);
    alarmTempTask.setActive(false);
}
