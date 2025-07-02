
#pragma once

#include <Arduino.h>
#include <../Communication/SerialManager.h>



#define MAX_PINS NUM_DIGITAL_PINS // SAME NUMBER SHARED BY ANALOG AND DIGITAL PINS, ANALOG PINS CAN BE USED AS DIGITAL PINS

/* --- TIME SERVICES ---*/

class ITimeKeeper {
protected:
    ITimeKeeper();
    unsigned long currentTime = 0;

public:

  virtual ~ITimeKeeper() = default; // default destructor for all derived classes
  static ITimeKeeper& getInstance();
  unsigned long getCurrentTime();
  virtual void update() = 0;

  //ITimeKeeper(const ITimeKeeper&) = delete;  // TimeKeeper tk2 = tk1;  // NO
  void operator=(const ITimeKeeper&) = delete; // tk2 = tk1; // NO
};


/* Class that uses Millis() */
class TimeKeeper : public ITimeKeeper {
private:
    TimeKeeper();

public:
    static ITimeKeeper& getInstance();
    void update() override;
};

/* Class that uses a Mock Time for tests */
class MockTimeKeeper : public ITimeKeeper {
private:
    MockTimeKeeper();

public:
    static ITimeKeeper& getInstance();
    void update() override;
    void setTime(unsigned long newTime);
};


/* --- INPUT SERVICES --- */

class IInputKeeper {
protected:
    IInputKeeper();

public:

  virtual ~IInputKeeper() = default; // default destructor for all derived classes
  static IInputKeeper& getInstance();
  virtual bool getDigitalPinState(unsigned int pin) = 0;
  virtual unsigned int getAnalogPinValue(unsigned int pin) = 0;

  IInputKeeper(const IInputKeeper&) = delete;  // IInputKeeper tk2 = tk1;  // NO

};

/* Class that uses digitalRead() and analogRead() */
class RealInputKeeper : public IInputKeeper {

    private:
        RealInputKeeper();
    public:
        static IInputKeeper& getInstance();
        bool getDigitalPinState(unsigned int pin) override;
        unsigned int getAnalogPinValue(unsigned int pin) override;
    
};


/* Class that uses a Mock array of inputs for tests */
class MockInputKeeper : public IInputKeeper {

    private:
        MockInputKeeper();
        int pins[MAX_PINS]; // using one array for both digital and analog pins, because analog pins start at 14 and digital pins at 0

    public:
        static IInputKeeper& getInstance();
        bool getDigitalPinState(unsigned int pin) override;
        unsigned int getAnalogPinValue(unsigned int pin) override;
        void setDigitalPinState(unsigned int pin, unsigned int state);
        void setAnalogPinValue(unsigned int pin, unsigned int value);
};


/* --- SERVICE LOCATOR --- */

class ServiceLocator {

    private:
        static ITimeKeeper* timeKeeper; 
        static IInputKeeper* inputKeeper;
        static SerialManager* serialManager;


    public:
        static void setTimeKeeperInstance(ITimeKeeper& newTimeKeeper){
            timeKeeper = &newTimeKeeper;
        }

        static ITimeKeeper& getTimeKeeperInstance(){
            return *timeKeeper;
        }


        static void setInputKeeperInstance(IInputKeeper& newInputKeeper){
            inputKeeper = &newInputKeeper;
        }

        static IInputKeeper& getInputKeeperInstance(){
            return *inputKeeper;
        }

        static void setSerialManagerInstance(SerialManager& newSerialManager){
            serialManager = &newSerialManager;
        }

        static SerialManager& getSerialManagerInstance(){
            return *serialManager;
        }

};

