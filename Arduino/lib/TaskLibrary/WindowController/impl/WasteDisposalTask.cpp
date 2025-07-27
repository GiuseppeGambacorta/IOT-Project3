#include "../api/WindowControllerTask.h"


enum windowManualCommand : byte
{
    NONE = 0,
    UP = 1,
    DOWN = 2
};

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

    temperature = serialManager.getvar(0);
    actualMode = serialManager.getvar(1);
    windowcommand = serialManager.getvar(2);
    serialManager.addVariableToSend((byte *)&manualButtonPressed, VarType::INT);
    serialManager.addVariableToSend((byte *)&actualWindowPosition, VarType::INT);
}
void WindowControllerTask::tick()
{
    state = (WindowManagerState) *actualMode;
    buttonTrigger.update(manualButton.isActive());

    if (buttonTrigger.isActive()){
        oldState = state;
        manualButtonPressed = 1;
    } 

    if (state != oldState){
        manualButtonPressed = 0;
    }
    
    actualWindowPosition = *temperature;
   
   
   //display.write(("Window Position: " + String(motor.getPosition())).c_str());

   switch (*actualMode)
    {
    case AUTOMATIC:
        motor.setPosition(*temperature);
        display.write("Automatic mode");
        break;

    case MANUAL:

        switch (*windowcommand)
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
        
        display.write("Manual mode");
    
       // display.write(("Temperature: " + String(*temperature)).c_str());
        break;
    default:
        display.write("Unknown mode");
        break;
    }

   
}

void WindowControllerTask::reset()
{
}
