

#include <Arduino.h>

#include "SchedulerLibrary.h"
#include "ArduinoStandardLibrary.h"
#include "Components.h"
#include "Tasks.h"


TemperatureSensor tempSensor = TemperatureSensor(A2, 206, 1023, 0); // A2 pin, max range 206, map range 1023, offset 0
DigitalInput manualButton = DigitalInput(2, 250);


Motor motor = Motor(3,0,90,0);
Display display = Display(0x27, 16, 2);


SerialManager& serialManager = ServiceLocator::getSerialManagerInstance();


SerialInputTask serialinputTask;
SerialOutputTask serialoutputTask;

InputTask inputTask(tempSensor, manualButton);
OutputTask outputTask(motor, display);

/*

StdExecTask stdExecTask(door, display, openButton, closeButton, ledGreen, ledRed, userDetector);
AlarmLevelTask alarmLevelTask(door, display, ledGreen, ledRed, levelDetector);
AlarmTempTask alarmTempTask(ledGreen,ledRed,display,door,tempSensor);
WasteDisposalTask wasteDisposalTask(stdExecTask, alarmLevelTask, alarmTempTask, levelDetector, tempSensor);

*/

Scheduler scheduler;


void setup() {
    serialManager.init();

    display.init();
    display.init();
    //serialManager.addDebugMessage("System started");
    
    scheduler.init(50);
    
  
    serialoutputTask.init(250);
    serialinputTask.init(500);

    inputTask.init(50);
    outputTask.init(100);
    /*

    wasteDisposalTask.init(100);
    alarmLevelTask.init(100);
    alarmTempTask.init(100);
    stdExecTask.init(100);

    */
    serialoutputTask.setActive(true);
    serialinputTask.setActive(true);

    inputTask.setActive(true);


    outputTask.setActive(true);
    /*
    wasteDisposalTask.setActive(true);
    alarmLevelTask.setActive(true);
    alarmTempTask.setActive(true);
    stdExecTask.setActive(true);

    */

    scheduler.addTask(&serialoutputTask);
    scheduler.addTask(&serialinputTask);

    scheduler.addTask(&inputTask);


    scheduler.addTask(&outputTask);
    /*

    scheduler.addTask(&wasteDisposalTask);
    scheduler.addTask(&alarmLevelTask); 
    scheduler.addTask(&alarmTempTask);
    scheduler.addTask(&stdExecTask);

    */
}

void loop() {
  scheduler.schedule();
  
}