#pragma once

#include <Arduino.h>
#include "Protocol.h"

class SerialManager
{
private:
    Register internalRegister;
    Protocol protocol;
    unsigned int baudRate;
    bool connectionEstablished = false;

    SerialManager(unsigned int baudRate) : protocol(internalRegister), baudRate(baudRate) {}

public:
    static SerialManager& getInstance(unsigned int baudRate = 9600) {
        static SerialManager* instance;
        if (instance == nullptr) {
            instance = new SerialManager(baudRate);
        }
        return *instance;
    }

    void operator=(SerialManager const&) = delete; // serial = serial1; NO

    void init()
    {
        Serial.begin(baudRate);
    }

    bool isSerialAvailable()
    {
        return Serial;
    }

    bool doHandshake()
    {
        return protocol.doHandshake();
    }

    bool isConnectionEstablished()
    {
        return protocol.isConnectionEstablished() && isSerialAvailable();
    }

    void addVariableToSend(byte *var, VarType varType)
    {
        internalRegister.addVariable(var, varType);
    }

    void addVariableToSend(String *string)
    {
        internalRegister.addVariable(string);
    }

    void addDebugMessage(const char *message)
    {
        internalRegister.addDebugMessage(message);
    }

    void addEventMessage(const char *message)
    {
        internalRegister.addEventMessage(message);
    }

    void sendData()
    {
        Serial.flush(); // wait for the transmission of outgoing serial data to complete, before sending new data
        protocol.sendinitCommunicationData();
        protocol.sendVariables();
        protocol.sendDebugMessages();
        protocol.sendEventMessages();

        internalRegister.resetDebugMessages();
        internalRegister.resetEventMessages();
    }

    void getData()
    {
        protocol.getData();
    }

    //for now only int are supported
    int16_t* getvar(unsigned int index)
    {
        return internalRegister.getIncomingDataHeader(index);
    }
};
