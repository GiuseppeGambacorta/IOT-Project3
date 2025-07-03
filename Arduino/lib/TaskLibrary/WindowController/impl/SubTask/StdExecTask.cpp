#include "../../api/subTask/StdExecTask.h"

#include "avr/sleep.h"
#include "EnableInterrupt.h"

StdExecTask ::StdExecTask(Door& door,
                          Display& display,
                          DigitalInput& openButton,
                          DigitalInput& closeButton,
                          DigitalOutput& ledGreen,
                          DigitalOutput& ledRed,
                          Pir& userDetector)
    : door(door),
      display(display),
      openButton(openButton),
      closeButton(closeButton),
      ledGreen(ledGreen),
      ledRed(ledRed),
      userDetector(userDetector),
      openTimer(TOpen),
      userTimer(TSleep),
      closeTimer(TClose){
        this->state = READY;
        this->userStatus = false;
}

void StdExecTask ::tick(){
    openButton.update();
    closeButton.update();
    switch (state)
    {
    case READY:
        execReady();
        break;
    case OPEN:
        execOpen();
        break;
    case SLEEP:
        execSleep();
        break;
    }
}

void StdExecTask ::homingReady(){

    ledGreen.turnOn();
    ledRed.turnOff(); 
    door.close();
    display.write("PRESS OPEN TO INSERT WASTE");
}

void StdExecTask ::execReady(){
    
    homingReady();

    userTimer.active(!userDetector.isDetected());
    if (userTimer.isTimeElapsed()) {
        state = SLEEP;
        userTimer.reset();
    }else if (openButton.isActive()){
        openTimer.active(true);
        state = OPEN;
    }
    
}

void StdExecTask ::homingOpen(){  
    door.open();
    display.write("PRESS CLOSE WHEN YOU'RE DONE");

}

void StdExecTask ::execOpen(){
    homingOpen();

    closeTimer.active(true);

    if (closeButton.isActive() || closeTimer.isTimeElapsed()){
        closeTimer.active(false);
        closeTimer.reset();
        state = READY;
    }
}
   


void StdExecTask ::homingSleep(){
    display.off();
}

void wakeUp(){
}

void StdExecTask ::execSleep(){
    homingSleep();
    set_sleep_mode(SLEEP_MODE_PWR_DOWN);
    sleep_enable();
    enableInterrupt(userDetector.getPin(), wakeUp, HIGH);
    sleep_mode();
    disableInterrupt(userDetector.getPin());
    sleep_disable();
    display.on();
    state = READY;
}



void StdExecTask ::reset(){
    this->userTimer.reset();
    this->closeTimer.reset();
    this->openTimer.reset();
    this->active = true;
    this->state = READY;
}