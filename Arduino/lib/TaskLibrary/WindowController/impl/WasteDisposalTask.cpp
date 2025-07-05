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

    temperature = serialManager.getvar(0);
    actualMode = serialManager.getvar(1);
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
    
    actualWindowPosition = state;
   
   
   //display.write(("Window Position: " + String(motor.getPosition())).c_str());
   display.write("ciao"); 
  
   switch (state)
    {
    case AUTOMATIC:
        motor.setPosition(*temperature);
        display.write("Automatic mode");
        break;

    case MANUAL:
        motor.setPosition(*temperature-10);
        display.write("Manual mode");
    
       // display.write(("Temperature: " + String(*temperature)).c_str());
        break;
    default:
    display.write("No Mode");
        break;
    }
   
}

void WindowControllerTask::reset()
{
}
