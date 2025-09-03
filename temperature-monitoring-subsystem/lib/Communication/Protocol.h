#pragma once
#include <Arduino.h>

enum class VarType : byte
{
    BYTE,
    INT,
    STRING,
    FLOAT,
};

enum class MessageType : byte
{
    VAR,
    DEBUG,
    EVENT,
};

struct DataHeader
{
    MessageType messageType;
    VarType varType;
    byte id;
    byte size;
    byte *data;
};

class Register
{
private:
    static const int MAX_VARIABLES = 10;
    DataHeader variables[MAX_VARIABLES];
    unsigned int variablesCount = 0;
    static const int MAX_DEBUG = 30;
    DataHeader debugMessage[MAX_DEBUG];
    unsigned int debugCount = 0;

    static const int MAX_EVENTS = 30;
    DataHeader eventMessage[MAX_EVENTS];
    unsigned int eventCount = 0;

    static const int NUMBER_OF_INCOMING_DATA = 2;
    int datiInArrivo[NUMBER_OF_INCOMING_DATA]; //for now only int are supported

public:
    Register();

    void addVariable(byte *var, VarType varType);
    void addVariable(String *string);
    void addDebugMessage(const char *message);
    void addEventMessage(const char *message);
    void updateStringLength(unsigned int index, String *string);
    DataHeader *getVariableHeader(unsigned int index);
    DataHeader *getDebugMessageHeader(unsigned int index);
    DataHeader *getEventMessageHeader(unsigned int index);
    int *getIncomingDataHeader(unsigned int index);
    unsigned int getVariableCount();
    unsigned int getDebugMessageCount();
    unsigned int getEventMessageCount();
    void resetDebugMessages();
    void resetEventMessages();
};



class Protocol
{
private:
    Register& internalRegister;
    bool connectionEstablished = false;

public:
    Protocol(Register& reg) : internalRegister(reg) {}

    void sendinitCommunicationData();
    void sendVariables();
    void sendDebugMessages();
    void sendEventMessages();
    bool doHandshake();
    bool isConnectionEstablished();
    void getData();
};