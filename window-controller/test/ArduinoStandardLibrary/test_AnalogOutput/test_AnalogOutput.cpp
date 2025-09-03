#include <unity.h>
#include "ArduinoStandardLibrary.h"


// NO MOCK FOR OUTPUT FOR NOW
#define MAX_VALUE 100


// execute before each test
void setUp(void) {
}


// execute after each test
void tearDown(void) {
}

void test_AnalogOutput(void) {
    AnalogOutput led(0, MAX_VALUE);

    TEST_ASSERT_EQUAL_INT(0, led.getValue());

    led.setValue(50);
    TEST_ASSERT_EQUAL_INT(50, led.getValue());

    led.setValue(300);
    TEST_ASSERT_EQUAL_INT(100, led.getValue());
    
}



void setup() {
    UNITY_BEGIN(); 
    RUN_TEST(test_AnalogOutput);
    UNITY_END(); 
}

void loop() {
}