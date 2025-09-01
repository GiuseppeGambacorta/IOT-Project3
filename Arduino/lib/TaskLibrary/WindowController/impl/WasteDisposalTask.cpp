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
      actualState(WindowManagerState::NORMAL)
{

    
    serialManager.addVariableToSend((byte *)&manualButtonPressed, VarType::INT);
    serialManager.addVariableToSend((byte *)&actualWindowPosition, VarType::INT);
}
void WindowControllerTask::tick()
{

    temperature = *serialManager.getvar(0);
    actualMode = WindowManagerMode(*serialManager.getvar(1));
    windowcommand = windowManualCommand(*serialManager.getvar(2));
    actualState = WindowManagerState(*serialManager.getvar(3));
    systemWindowPos = *serialManager.getvar(4);


    buttonTrigger.update(manualButton.isActive());

    if (buttonTrigger.isActive()){
        oldMode = actualMode;
        manualButtonPressed = 1;
    } 

    if (actualMode != oldMode){
        manualButtonPressed = 0;
    }
    
    char msg[4 * 20];

    const char* modeStr;
    switch (actualMode)
    {
    case AUTOMATIC:
        modeStr = "Automatic";
        motor.setPosition(systemWindowPos);
        snprintf(msg, sizeof(msg), "Position:%d\nModality:%s\0", motor.getPosition(), modeStr);
        break;
    case MANUAL:
        modeStr = "Manual";

        if (windowcommand != oldCommand) {
            switch (windowcommand)
            {
            case UP:
                motor.setPosition(motor.getPosition() + 5);
                break;
            case DOWN:
                motor.setPosition(motor.getPosition() - 5);
                break;
            default:
                motor.setPosition(motor.getPosition());
                break;
            }
            oldCommand = windowcommand;
    }
        snprintf(msg, sizeof(msg), "Position:%d\nModality:%s\nTemperature:%d\0", motor.getPosition(), modeStr, temperature);
        break;
    default:
        modeStr = "Unknown";
        snprintf(msg, sizeof(msg), "Position:%d\nModality:%s\0", motor.getPosition(), modeStr);
        break;
    }
        display.write(msg);
        actualWindowPosition =motor.getPosition();


}

void WindowControllerTask::reset()
{
}
