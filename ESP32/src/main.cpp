#include <Arduino.h>
#include "ArduinoStandardLibrary.h"
#include <TemperatureSensor.h>



TemperatureSensor tempSensor = TemperatureSensor(A2);
DigitalOutput Redled = DigitalOutput(1);
DigitalOutput Greenled = DigitalOutput(2);


InputTask inputTask(levelDetector, userDetector, tempSensor, openButton, closeButton);
OutputTask outputTask(door, display ,ledGreen, ledRed);

Scheduler scheduler;


void setup() {
    
    scheduler.init(50);
    


    inputTask.init(50);
  
    outputTask.init(100);



    inputTask.setActive(true);
 
    outputTask.setActive(true);
   


    scheduler.addTask(&inputTask);
    scheduler.addTask(&outputTask);
}

void loop() {
  scheduler.schedule();
  
}