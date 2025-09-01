#include <Arduino.h>
#include "Components.h"

// Tutto il resto commentato
/*
#include "SchedulerLibrary.h"
#include "ArduinoStandardLibrary.h"
#include "Tasks.h"

AnalogInput potentiomenter = AnalogInput(A2, 90, 1023, 0); // A2 pin, max range 206, map range 1023, offset 0
DigitalInput manualButton = DigitalInput(4, 250);

Motor motor = Motor(3, 0, 90, 0);
Display display = Display(0x27, 16, 2);



SerialInputTask serialinputTask;
SerialOutputTask serialoutputTask;

InputTask inputTask(potentiomenter, manualButton);
OutputTask outputTask(motor, display);

WindowControllerTask windowController(potentiomenter, manualButton, motor, display);

Scheduler scheduler;
*/
SerialManager &serialManager = ServiceLocator::getSerialManagerInstance();
Display display = Display(0x27, 16, 2);
DigitalOutput led = DigitalOutput(5);
void setup()
{
  serialManager.init();
  display.init();
}


void loop()
{
  int16_t values[5] = {0};

  if (!serialManager.isConnectionEstablished()){
  serialManager.doHandshake();
  led.turnOff();
  } else
  {
    serialManager.getData();
    led.turnOn();
    for (int i =0; i < 5; i++){
      values[i] = *serialManager.getvar(i);
    }
  }




     char msg[33];
    snprintf(msg, sizeof(msg), "%d %d %d %d %d", values[0], values[1], values[2], values[3], values[4]);
    display.write(msg);
    display.update();
  
 led.update();
  
}