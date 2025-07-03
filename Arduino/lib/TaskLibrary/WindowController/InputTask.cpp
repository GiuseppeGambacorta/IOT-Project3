#include "../../api/subTask/InputTask.h"

InputTask::InputTask(
                 Sonar& levelDetector,
                 Pir& userDetector,
                 TemperatureSensor& tempSensor,
                 DigitalInput& openButton, 
                 DigitalInput& closeButton) 
    : 
    levelDetector(levelDetector),
    userDetector(userDetector),
    tempSensor(tempSensor),
    openButton(openButton), 
    closeButton(closeButton) {
    ServiceLocator::getSerialManagerInstance().addVariableToSend((byte *)&temp, VarType::INT);
    ServiceLocator::getSerialManagerInstance().addVariableToSend((byte *)&level, VarType::FLOAT);
}

void InputTask::tick() {
    userDetector.update();
    levelDetector.update();
    tempSensor.update();
    openButton.update();
    closeButton.update();

    temp = tempSensor.readTemperature();
    level = levelDetector.readDistance();
}

void InputTask::reset() {
    
}