#include <unity.h>
#include "ArduinoStandardLibrary.h"

#define MAX_MAP_VALUE 100
#define ANALOG_PIN A0

// execute before each test
void setUp(void) {
    ServiceLocator::setInputKeeperInstance(MockInputKeeper::getInstance());
    ServiceLocator::setTimeKeeperInstance(MockTimeKeeper::getInstance());
}


// execute after each test
void tearDown(void) {
}


void test_AnalogInput_MapAndFilter(void) {
    MockInputKeeper& InputKeeper = (MockInputKeeper&) ServiceLocator::getInputKeeperInstance();
    AnalogInput potentiometer(ANALOG_PIN,MAX_MAP_VALUE);

    TEST_ASSERT_EQUAL_INT(0, potentiometer.getValue());
    
    InputKeeper.setAnalogPinValue(ANALOG_PIN, 0);
    potentiometer.update();
    TEST_ASSERT_EQUAL_INT(0, potentiometer.getValue());

    InputKeeper.setAnalogPinValue(ANALOG_PIN, 1023);
    potentiometer.update();
    TEST_ASSERT_EQUAL_INT(50, potentiometer.getValue());


    InputKeeper.setAnalogPinValue(ANALOG_PIN, 512);
    potentiometer.update();
    TEST_ASSERT_EQUAL_INT(50, potentiometer.getValue());

    InputKeeper.setAnalogPinValue(ANALOG_PIN, 1023 - 102);
    potentiometer.update();
    TEST_ASSERT_EQUAL_INT(60, potentiometer.getValue());  // (0+ 100 + 50  +90)/ 4 = 60
      
}



void setup() {
    UNITY_BEGIN(); 
    RUN_TEST(test_AnalogInput_MapAndFilter);
    UNITY_END(); 
}

void loop() {
}