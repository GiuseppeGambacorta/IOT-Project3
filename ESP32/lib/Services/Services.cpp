#include "Services.h"


/*---- SERVICE LOCATOR ----*/

// allocation of the static variable
ITimeKeeper* ServiceLocator::timeKeeper = &TimeKeeper::getInstance();
IInputKeeper* ServiceLocator::inputKeeper = &RealInputKeeper::getInstance();
SerialManager* ServiceLocator::serialManager = &SerialManager::getInstance(9600);


/*---- TIME KEEPER ABSTRACT CLASS ----*/

ITimeKeeper::ITimeKeeper()  {}

unsigned long ITimeKeeper::getCurrentTime() {
    return this->currentTime;
}


/*---- TIME KEEPER WITH MILLIS() ----*/

TimeKeeper::TimeKeeper() : ITimeKeeper() {}

ITimeKeeper& TimeKeeper::getInstance() {
    static TimeKeeper instance;
    return instance;
}

void TimeKeeper::update() {
    this->currentTime = millis();
}



/*---- MOCK TIME KEEPER ----*/

MockTimeKeeper::MockTimeKeeper() : ITimeKeeper() {}

ITimeKeeper& MockTimeKeeper::getInstance() {
    static MockTimeKeeper instance;
    return instance;
}

void MockTimeKeeper::update() {
    ;
}

void MockTimeKeeper::setTime(unsigned long newTime) {
    this->currentTime = newTime;
}





/*---- INPUT KEEPER ABSTRACT CLASS ----*/

IInputKeeper::IInputKeeper() {}


/*---- INPUT KEEPER WITH DIGITALREAD() AND ANALOGREAD() ----*/


RealInputKeeper::RealInputKeeper() : IInputKeeper() {}

IInputKeeper& RealInputKeeper::getInstance() {
    static RealInputKeeper instance;
    return instance;
}

bool RealInputKeeper::getDigitalPinState(unsigned int pin) {

    if (pin >= NUM_DIGITAL_PINS) {
        return false;
    }
    return digitalRead(pin);
}

unsigned int RealInputKeeper::getAnalogPinValue(unsigned int pin) {

    if (pin >= NUM_DIGITAL_PINS) {
        return 0;
    }
    return analogRead(pin);
}


/* ---- MOCK INPUT KEEPER ---- */

MockInputKeeper::MockInputKeeper() : IInputKeeper() {}

IInputKeeper& MockInputKeeper::getInstance() {
    static MockInputKeeper instance;
    return instance;
}


bool MockInputKeeper::getDigitalPinState(unsigned int pin) {
    if (pin < MAX_PINS) {
        return pins[pin];
    }
    return false;
}

unsigned int MockInputKeeper::getAnalogPinValue(unsigned int pin) {
    if (pin < MAX_PINS) {
        return pins[pin];
    }
    return 0;
}



void MockInputKeeper::setDigitalPinState(unsigned int pin, unsigned int state) {
    if (pin < MAX_PINS) {
        pins[pin] = state;
    }
}


void MockInputKeeper::setAnalogPinValue(unsigned int pin, unsigned int value) {
    if (pin < MAX_PINS) {
        pins[pin] = value;
    }
}
