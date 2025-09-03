#include "../api/Door.h"

Door::Door(unsigned int pin) : motor(pin, 90, 90, -90) {
}


void Door::init() {
    motor.init();
}

void Door::open() {
    motor.setPosition(90);
}

void Door::close() {
    motor.setPosition(0);
}

void Door::empty() {
    motor.setPosition(-90);
}

void Door::update() {
    motor.update();
}


bool Door::isClosed() {
    return motor.getPosition() == 0  && motor.isInPosition() ;
}

bool Door::isOpened() {
    return motor.getPosition() == 90  && motor.isInPosition();
}

bool Door::isInEmptyPosition() {
    return motor.getPosition() == -90  && motor.isInPosition();
}