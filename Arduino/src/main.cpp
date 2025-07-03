

#include <Arduino.h>

#include "SchedulerLibrary.h"
#include "ArduinoStandardLibrary.h"
#include "Components.h"
#include "Tasks.h"

AnalogInput potentiomenter = AnalogInput(A2, 90, 1023, 0); // A2 pin, max range 206, map range 1023, offset 0
DigitalInput manualButton = DigitalInput(4, 250);

Motor motor = Motor(3, 0, 90, 0);
Display display = Display(0x27, 16, 2);

SerialManager &serialManager = ServiceLocator::getSerialManagerInstance();

SerialInputTask serialinputTask;
SerialOutputTask serialoutputTask;

InputTask inputTask(potentiomenter, manualButton);
OutputTask outputTask(motor, display);

WindowControllerTask windowController(potentiomenter, manualButton, motor, display);

Scheduler scheduler;

void setup()
{
  serialManager.init();

  display.init();
  motor.init();
  // serialManager.addDebugMessage("System started");

  scheduler.init(50);

  serialoutputTask.init(250);
  serialinputTask.init(500);
  inputTask.init(100);
  windowController.init(100);
  outputTask.init(100);

  serialoutputTask.setActive(true);
  serialinputTask.setActive(true);
  inputTask.setActive(true);
  windowController.setActive(true);
  outputTask.setActive(true);

  scheduler.addTask(&serialoutputTask);
  scheduler.addTask(&serialinputTask);
  scheduler.addTask(&inputTask);
  scheduler.addTask(&windowController);
  scheduler.addTask(&outputTask);
}

void loop()
{
 // display.write("System running");
 // display.update();
  scheduler.schedule();
}