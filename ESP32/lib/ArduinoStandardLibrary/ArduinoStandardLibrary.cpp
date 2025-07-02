#include "ArduinoStandardLibrary.h"

/* ---- SCHEDULER TIMER ---- */

SchedulerTimer::SchedulerTimer() {}

void SchedulerTimer::calculateNextTick() {
    nextTickTime = timeKeeper.getCurrentTime() + tickInterval;
}

void SchedulerTimer::setupFreq(int freq) {
    tickInterval = 1000 / freq; // Converti frequenza in intervallo (ms)
    calculateNextTick();
}

void SchedulerTimer::setupPeriod(int period) {
    tickInterval = period;
    calculateNextTick();
}

void SchedulerTimer::waitForNextTick() {
    while (true) {
        timeKeeper.update();
        unsigned long currentTime = timeKeeper.getCurrentTime();
        if (currentTime >= nextTickTime) {
            calculateNextTick();
            break;
        }
    }
}

/* ---- TIMER ---- */

Timer::Timer(unsigned long timeDuration) : timeDuration(timeDuration) {}

void Timer::active(bool start) {
    if (!start) {
        startInterlock = false;
    }

    if (!this->startInterlock) {
        if (start) {
            startInterlock = true;
            oldTime = timeKeeper.getCurrentTime();
        }
    }
}

bool Timer::isTimeElapsed() {
    if (this->startInterlock) {
        unsigned long currentTime = timeKeeper.getCurrentTime();
        if (currentTime - this->oldTime >= this->timeDuration) {
            return true;
        }
    }
    return false;
}

void Timer::setTime(unsigned long newTime) {
    this->timeDuration = newTime;
}

void Timer::reset() {
    this->oldTime = 0;
    this->startInterlock = 0;
}

/* ---- DIGITAL INPUT ---- */

DigitalInput::DigitalInput(unsigned int pin, unsigned long threshold) : pin(pin) {
    pinMode(pin, INPUT);
    this->activationTimer = new Timer(threshold);
}

void DigitalInput::update() {
    this->activationTimer->active(this->inputKeeper.getDigitalPinState(this->pin));
    this->value = this->activationTimer->isTimeElapsed();
    this->trigChanged = this->value != this->oldValue;
    this->oldValue = this->value;
}

bool DigitalInput::isActive() {
    return this->value;
}

bool DigitalInput::isChanged() {
    return this->trigChanged;
}

/* ---- DIGITAL OUTPUT ---- */

DigitalOutput::DigitalOutput(unsigned int pin) : pin(pin) {
    pinMode(pin, OUTPUT);
}

void DigitalOutput::update() {
    digitalWrite(this->pin, this->value);
}

void DigitalOutput::turnOn() {
    this->value = 1;
}

void DigitalOutput::turnOff() {
    this->value = 0;
}

bool DigitalOutput::isActive() {
    return this->value;
}

/* ---- ANALOG INPUT ---- */

AnalogInput::AnalogInput(unsigned int pin, unsigned int mapValue) : pin(pin), mapValue(mapValue) {
    pinMode(pin, INPUT);
}

float AnalogInput::filterValue(unsigned int inputValue) {
    array[currentIndex] = inputValue;
    currentIndex = (currentIndex + 1) % maxFilterSize;
    if (filterCount < maxFilterSize) {
        filterCount++;
    }

    unsigned long sum = 0;
    for (unsigned int i = 0; i < filterCount; i++) {
        sum += array[i];
    }
    int returnValue = sum / filterCount;

    return returnValue;
}

void AnalogInput::update() {
    int actualValue = this->inputKeeper.getAnalogPinValue(this->pin);
    this->value = this->filterValue(map(actualValue, 0, 1023, 0, this->mapValue));
}

int AnalogInput::getValue() {
    return this->value;
}

/* ---- ANALOG OUTPUT ---- */

AnalogOutput::AnalogOutput(unsigned int pin, unsigned int maxValue) : pin(pin), maxValue(maxValue) {
    pinMode(pin, OUTPUT);
}

void AnalogOutput::setValue(unsigned int value) {
    if (value <= this->maxValue) {
        this->value = value;
    } else {
        this->value = this->maxValue;
    }
}

int AnalogOutput::getValue() {
    return this->value;
}

void AnalogOutput::update() {
    analogWrite(this->pin, map(this->value, 0, this->maxValue, 0, 255));
}