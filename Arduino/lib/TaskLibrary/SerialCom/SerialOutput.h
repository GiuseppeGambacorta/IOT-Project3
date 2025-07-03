#pragma once

#include <../Services/Services.h>
#include "./TaskLibrary.h"


class SerialOutputTask : public Task {

private:
    SerialManager& serialManager;
    
public:
    SerialOutputTask() : serialManager(ServiceLocator::getSerialManagerInstance()) {}

    void tick() override {
        if (!serialManager.isConnectionEstablished()){
            serialManager.doHandshake();
        } else {
            serialManager.sendData();
        }
    }
    void reset() override {
        ;
    }
};
