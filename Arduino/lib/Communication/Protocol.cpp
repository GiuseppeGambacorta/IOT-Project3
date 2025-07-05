#include "Protocol.h"

Register::Register() {}

void Register::addVariable(byte *var, VarType varType)
{
    if (variablesCount < MAX_VARIABLES)
    {
        variables[variablesCount].messageType = MessageType::VAR;
        variables[variablesCount].varType = varType;
        variables[variablesCount].id = variablesCount;
        variables[variablesCount].data = var;
        switch (varType)
        {
        case VarType::BYTE:
            variables[variablesCount].size = sizeof(byte);
            break;
        case VarType::INT:
            variables[variablesCount].size = sizeof(int);
            break;
        case VarType::FLOAT:
            variables[variablesCount].size = sizeof(float);
            break;
        default:
            break; // for now only byte and int are supported
        }
        variablesCount++;
    }
}

void Register::addVariable(String *string)
{
    if (variablesCount < MAX_VARIABLES)
    {
        variables[variablesCount].messageType = MessageType::VAR;
        variables[variablesCount].varType = VarType::STRING;
        variables[variablesCount].id = variablesCount;
        variables[variablesCount].data = (byte *)string;
        variables[variablesCount].size = string->length() + 1;
        variablesCount++;
    }
}

void Register::addDebugMessage(const char *message)
{
    for (unsigned int i = 0; i < debugCount; i++)
    {
        if (strcmp((char *)debugMessage[i].data, message) == 0)
        {
            return;
        }
    }

    if (debugCount < MAX_DEBUG)
    {
        debugMessage[debugCount].messageType = MessageType::DEBUG;
        debugMessage[debugCount].varType = VarType::STRING;
        debugMessage[debugCount].id = 0;
        debugMessage[debugCount].data = (byte *)message;
        debugMessage[debugCount].size = strlen(message) + 1;
        debugCount++;
    }
}

void Register::addEventMessage(const char *message)
{
    for (unsigned int i = 0; i < eventCount; i++)
    {
        if (strcmp((char *)eventMessage[i].data, message) == 0)
        {
            return;
        }
    }

    if (eventCount < MAX_EVENTS)
    {
        eventMessage[eventCount].messageType = MessageType::EVENT;
        eventMessage[eventCount].varType = VarType::STRING;
        eventMessage[eventCount].id = 0;
        eventMessage[eventCount].data = (byte *)message;
        eventMessage[eventCount].size = strlen(message) + 1;
        eventCount++;
    }
}

void Register::updateStringLength(unsigned int index, String *string)
{
    if (index >= 0 && index < variablesCount && variables[index].varType == VarType::STRING)
    {
        variables[index].size = string->length() + 1;
    }
}

DataHeader *Register::getVariableHeader(unsigned int index)
{
    if (index >= 0 && index < variablesCount)
    {
        return &variables[index];
    }
    else
    {
        return nullptr;
    }
}

DataHeader *Register::getDebugMessageHeader(unsigned int index)
{
    if (index >= 0 && index < variablesCount)
    {
        return &debugMessage[index];
    }
    else
    {
        return nullptr;
    }
}

DataHeader *Register::getEventMessageHeader(unsigned int index)
{
    if (index >= 0 && index < variablesCount)
    {
        return &eventMessage[index];
    }
    else
    {
        return nullptr;
    }
}

int *Register::getIncomingDataHeader(unsigned int index)
{
    if (index >= 0 && index < NUMBER_OF_INCOMING_DATA)
    {
        return &datiInArrivo[index];
    }
    else
    {
        return nullptr;
    }
}

unsigned int Register::getVariableCount()
{
    return variablesCount;
}

unsigned int Register::getDebugMessageCount()
{
    return debugCount;
}

unsigned int Register::getEventMessageCount()
{
    return eventCount;
}

void Register::resetDebugMessages()
{
    debugCount = 0;
}

void Register::resetEventMessages()
{
    eventCount = 0;
}

void Protocol::sendinitCommunicationData()
{
    byte numberOfVariables = internalRegister.getVariableCount() + internalRegister.getDebugMessageCount() + internalRegister.getEventMessageCount();
    byte header = 255;
    Serial.write(header);
    header = 0;
    Serial.write(header);
    Serial.write((byte *)&numberOfVariables, sizeof(numberOfVariables));
}

void Protocol::sendVariables()
{
    int numberOfVariables = internalRegister.getVariableCount();
    for (unsigned int i = 0; i < numberOfVariables; i++)
    {
        DataHeader *header = internalRegister.getVariableHeader(i);

        Serial.write((byte *)&header->messageType, sizeof(header->messageType));
        Serial.write((byte *)&header->varType, sizeof(header->varType));
        Serial.write((byte *)&header->id, sizeof(header->id));

        if (header->varType == VarType::STRING)
        {
            internalRegister.updateStringLength(i, (String *)header->data);
            Serial.write((byte *)&header->size, sizeof(header->size));
            String *string = (String *)header->data;
            Serial.write(string->c_str(), header->size);
        }
        else
        {
            Serial.write((byte *)&header->size, sizeof(header->size));
            Serial.write(header->data, header->size);
        }
    }
}

void Protocol::sendDebugMessages()
{
    for (unsigned int i = 0; i < internalRegister.getDebugMessageCount(); i++)
    {
        DataHeader *header = internalRegister.getDebugMessageHeader(i);

        Serial.write((byte *)&header->messageType, sizeof(header->messageType));
        Serial.write((byte *)&header->varType, sizeof(header->varType));
        Serial.write((byte *)&header->id, sizeof(header->id));
        Serial.write((byte *)&header->size, sizeof(header->size));
        Serial.write(header->data, header->size);
    }
}

void Protocol::sendEventMessages()
{
    for (unsigned int i = 0; i < internalRegister.getEventMessageCount(); i++)
    {
        DataHeader *header = internalRegister.getEventMessageHeader(i);

        Serial.write((byte *)&header->messageType, sizeof(header->messageType));
        Serial.write((byte *)&header->varType, sizeof(header->varType));
        Serial.write((byte *)&header->id, sizeof(header->id));
        Serial.write((byte *)&header->size, sizeof(header->size));
        Serial.write(header->data, header->size);
    }
}

bool Protocol::doHandshake()
{
    if (Serial.available() > 0)
    {
        byte received = (short unsigned int)Serial.read(); //convert because i want to check a number not a char or a byte
        if (received == 255)
        {
            Serial.write(10);
            connectionEstablished = true;
        }
    }
    return connectionEstablished;
}

bool Protocol::isConnectionEstablished()
{
    return connectionEstablished;
}

void Protocol::getData()
{
    if (Serial.available() > 0)
    {
        byte header = Serial.read();
        if (header == 255)
        {
            byte command = Serial.read();
            if (command == 0)
            {

                byte id = Serial.read();
                byte size = Serial.read();
                byte buffer[size];
                Serial.readBytes(buffer, size);
                int *var = internalRegister.getIncomingDataHeader(int(id));
                if (var != nullptr)
                {
                    *var = (int(buffer[1]) << 8) | int(buffer[0]); //for now only int are supported
                }
            }
        }
    }
}


