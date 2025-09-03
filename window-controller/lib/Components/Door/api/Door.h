#pragma once
#include "../../Motor/api/Motor.h"

class Door {
private:
    Motor motor;

public:
    Door(unsigned int pin);
    void init();
    void open();
    void close();
    void empty();
    bool isClosed();
    bool isOpened();
    bool isInEmptyPosition();
    void update();
};