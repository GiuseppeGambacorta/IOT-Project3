#pragma once


#include <Arduino.h>
#include "Services.h"

/* ---- SCHEDULER TIMER ---- */

class SchedulerTimer {
private:
    ITimeKeeper& timeKeeper = ServiceLocator::getTimeKeeperInstance();
    unsigned long tickInterval = 0;
    unsigned long nextTickTime = 0;

    void calculateNextTick(); // Calcola il tempo del prossimo tick

public:
    SchedulerTimer();
    void setupFreq(int freq);     // Configura la frequenza (Hz)
    void setupPeriod(int period); // Configura il periodo (ms)
    void waitForNextTick();       // Attende il prossimo tick
};

/* ---- TIMER ---- */

class Timer {
private:
    unsigned long timeDuration;
    unsigned long oldTime = 0;
    bool startInterlock = false;
    ITimeKeeper& timeKeeper = ServiceLocator::getTimeKeeperInstance();

public:
    Timer(unsigned long timeDuration);

    void active(bool start);
    bool isTimeElapsed();
    void setTime(unsigned long newTime);
    void reset();
};

/* ---- DIGITAL INPUT ---- */

class DigitalInput {
private:
    unsigned int pin;
    Timer* activationTimer;
    unsigned int value = 0;
    unsigned int oldValue = 0;
    unsigned int trigChanged = 0;
    IInputKeeper& inputKeeper = ServiceLocator::getInputKeeperInstance();

public:
    DigitalInput(unsigned int pin, unsigned long threshold);

    void update();
    bool isActive();
    bool isChanged();
};

/* ---- DIGITAL OUTPUT ---- */

class DigitalOutput {
private:
    unsigned int pin;
    unsigned int value = 0;

public:
    DigitalOutput(unsigned int pin);

    void update();
    void turnOn();
    void turnOff();
    bool isActive();
};

/* ---- ANALOG INPUT ---- */

class AnalogInput {
private:
    unsigned int pin;
    unsigned int mapValue;
    float value = 0;
    static const unsigned int maxFilterSize = 10;
    int array[maxFilterSize] = {0};
    unsigned int filterCount = 0;
    float val_coef = 0;
    int currentIndex = 0;
    float filterValue(unsigned int inputValue);
    IInputKeeper& inputKeeper = ServiceLocator::getInputKeeperInstance();

public:
    AnalogInput(unsigned int pin, unsigned int mapValue);
    void update();
    int getValue();
};

/* ---- ANALOG OUTPUT ---- */

class AnalogOutput {
private:
    unsigned int pin;
    unsigned int maxValue;
    unsigned int value = 0;

public:
    AnalogOutput(unsigned int pin, unsigned int maxValue);

    void setValue(unsigned int value);
    int getValue();
    void update();
};
