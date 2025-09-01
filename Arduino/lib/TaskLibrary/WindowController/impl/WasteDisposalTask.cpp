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
        oldState = actualState;
        manualButtonPressed = 1;
    } 

    if (actualState != oldState){
        manualButtonPressed = 0;
    }
    

   char msg[32]; // Buffer per il messaggio

    const char* modeStr;
    switch (actualMode)
    {
    case AUTOMATIC:
        modeStr = "Auto";
        motor.setPosition(systemWindowPos);
        break;
    case MANUAL:
        modeStr = "Manual";
        switch (windowcommand)
        {
        case UP:
            motor.setPosition(motor.getPosition() + 5);
            break;
        case DOWN:
            motor.setPosition(motor.getPosition() - 5);
            break;
        case NONE:
            motor.setPosition(motor.getPosition());
            break;
        default:
            break;
        }
        break;
    default:
        modeStr = "Sconosciuta";
        motor.setPosition(systemWindowPos);
        break;
    }
    //actualState = WindowManagerState(56);
    snprintf(msg, sizeof(msg), "Posizione:%d, Mod:%d", motor.getPosition(), WindowManagerState(*serialManager.getvar(3)));
    display.write(msg);

    // ...existing code...
}

void WindowControllerTask::reset()
{
}
