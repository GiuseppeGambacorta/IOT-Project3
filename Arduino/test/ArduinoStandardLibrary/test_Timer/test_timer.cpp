#include <unity.h>
#include "ArduinoStandardLibrary.h"


// execute before each test
void setUp(void) {
    ServiceLocator::setTimeKeeperInstance(MockTimeKeeper::getInstance());
}


// execute after each test
void tearDown(void) {
}

void test_timer(void) {
    MockTimeKeeper& TimeKeeper = (MockTimeKeeper&) ServiceLocator::getTimeKeeperInstance();
    Timer timer(1000);
    TEST_ASSERT_FALSE(timer.isTimeElapsed());

    timer.active(true);
    TimeKeeper.setTime(900);
    TEST_ASSERT_FALSE(timer.isTimeElapsed());

    TimeKeeper.setTime(1000);
    TEST_ASSERT_TRUE(timer.isTimeElapsed());

    TimeKeeper.setTime(1100);
    TEST_ASSERT_TRUE(timer.isTimeElapsed());

    timer.active(false);
    TEST_ASSERT_FALSE(timer.isTimeElapsed());

    timer.active(true);
    TEST_ASSERT_FALSE(timer.isTimeElapsed());

    TimeKeeper.setTime(2100);
    TEST_ASSERT_TRUE(timer.isTimeElapsed());
}



void setup() {
    UNITY_BEGIN(); 
    RUN_TEST(test_timer);
    UNITY_END(); 
}

void loop() {
   
}