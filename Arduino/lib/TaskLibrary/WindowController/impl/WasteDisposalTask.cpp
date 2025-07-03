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

    if (manualButton.isActive() == 1)
    {
        state = WindowManagerState::MANUAL;
    }
    else
    {
        state = WindowManagerState::AUTOMATIC;
    }

   
   // display.clear();
   //display.write(("Window Position: " + String(motor.getPosition())).c_str());
   display.write("ciao"); 
   switch (state)
    {
    case AUTOMATIC:
        motor.setPosition(90);
        display.write("Automatic mode");
        break;

    case MANUAL:
        motor.setPosition(67);
        display.write("Manual mode");
    
       // display.write(("Temperature: " + String(*temperature)).c_str());
        break;
    default:
        break;
    }
}

void WindowControllerTask::reset()
{
}
