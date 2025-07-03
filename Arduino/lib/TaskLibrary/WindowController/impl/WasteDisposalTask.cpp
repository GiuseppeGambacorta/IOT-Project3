#include "../api/WindowControllerTask.h"


WindowControllerTask::WindowControllerTask(
    AnalogInput &potentiometer,
    DigitalInput &manualButton,
    Motor &motor,
    Display &display)
    : potentiometer(potentiometer),
      manualButton(manualButton),
      motor(motor),
      display(display),
      state(WindowManagerState::AUTOMATIC)
{

    temperature = ServiceLocator::getSerialManagerInstance().getvar(0);
    actualMode = ServiceLocator::getSerialManagerInstance().getvar(1);
}
void WindowControllerTask::tick()
{
    double level = levelDetector.readDistance();
    int temp = tempSensor.readTemperature();
    tempTimer.active(temp >= maxTemp);

    switch (state)
    {
    case STD_EXEC:

        if (level <= maxLevel)
        {
            state = WasteDisposalState::LVL_ALLARM;
        }

        if (tempTimer.isTimeElapsed())
        {
            state = WasteDisposalState::TEMP_ALLARM;
        }
        break;
    case LVL_ALLARM:
        if (*empty == ResetMessage::MESSAGE_SEEN)
        {
            state = WasteDisposalState::STD_EXEC;
            *empty = ResetMessage::NO_MESSAGE;
        }

        if (tempTimer.isTimeElapsed())
        {
            state = WasteDisposalState::TEMP_ALLARM;
        }
        break;
    case TEMP_ALLARM:

        if (*fire == ResetMessage::MESSAGE_SEEN)
        {
            state = WasteDisposalState::STD_EXEC;
            *fire = ResetMessage::NO_MESSAGE;
            ;
        }
    }

    switch (state)
    {
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

    if (state != oldState)
    {
        ServiceLocator::getSerialManagerInstance().addEventMessage(wasteDisposalStateToString(state));
    }
    oldState = state;
}

void WindowControllerTask::reset()
{
    
}
