#include <unity.h>
#include "ArduinoStandardLibrary.h"


// NO MOCK FOR OUTPUT FOR NOW

// execute before each test
void setUp(void) {
}


// execute after each test
void tearDown(void) {
}

void test_DigitalOutput(void) {
    DigitalOutput led(0);

    TEST_ASSERT_FALSE(led.isActive());

    led.turnOn();
    TEST_ASSERT_TRUE(led.isActive());

    led.turnOff();
    TEST_ASSERT_FALSE(led.isActive());
    
}



void setup() {
    UNITY_BEGIN(); 
    RUN_TEST(test_DigitalOutput);
    UNITY_END(); 
}

void loop() {
}