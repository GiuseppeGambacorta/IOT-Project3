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

    if (*actualMode == 1)
    {
        state = WindowManagerState::MANUAL;
    }
    else
    {
        state = WindowManagerState::AUTOMATIC;
    }

    display.on();
    display.clear();
    display.write(("Window Position: " + String(motor.getPosition())).c_str());
    switch (state)
    {
    case AUTOMATIC:
        motor.setPosition(50);
        display.write("Automatic mode");
        break;

    case MANUAL:
        motor.setPosition(potentiometer.getValue());
        display.write("Manual mode");
        display.write(("Temperature: " + String(*temperature)).c_str());
        break;
    default:
        break;
    }
}

void WindowControllerTask::reset()
{
}
