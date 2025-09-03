#include <unity.h>
#include "ArduinoStandardLibrary.h"


// execute before each test
void setUp(void) {
    ServiceLocator::setInputKeeperInstance(MockInputKeeper::getInstance());
    ServiceLocator::setTimeKeeperInstance(MockTimeKeeper::getInstance());
}


// execute after each test
void tearDown(void) {
}

void test_DigitalInput_NoDelay(void) {
    MockInputKeeper& InputKeeper = (MockInputKeeper&) ServiceLocator::getInputKeeperInstance();
    DigitalInput button(0,0);

    TEST_ASSERT_FALSE(button.isActive());
    TEST_ASSERT_FALSE(button.isChanged());

    InputKeeper.setDigitalPinState(0, true);
    button.update();
    TEST_ASSERT_TRUE(button.isActive());
    TEST_ASSERT_TRUE(button.isChanged());

    button.update();
    TEST_ASSERT_TRUE(button.isActive());
    TEST_ASSERT_FALSE(button.isChanged());

    InputKeeper.setDigitalPinState(0, false);
    button.update();

    TEST_ASSERT_FALSE(button.isActive());
    TEST_ASSERT_TRUE(button.isChanged());
 
}

void test_digitalInput_WithDelay(void) {
    MockInputKeeper& InputKeeper = (MockInputKeeper&) ServiceLocator::getInputKeeperInstance();
    MockTimeKeeper& TimeKeeper = (MockTimeKeeper&) ServiceLocator::getTimeKeeperInstance();
    DigitalInput button(0,1000);

    TEST_ASSERT_FALSE(button.isActive());
    TEST_ASSERT_FALSE(button.isChanged());

    InputKeeper.setDigitalPinState(0, true);
    button.update();
    TEST_ASSERT_FALSE(button.isActive());
    TEST_ASSERT_FALSE(button.isChanged());

    TimeKeeper.setTime(900);
    button.update();
    TEST_ASSERT_FALSE(button.isActive());
    TEST_ASSERT_FALSE(button.isChanged());

    TimeKeeper.setTime(1000);
    button.update();
    TEST_ASSERT_TRUE(button.isActive());
    TEST_ASSERT_TRUE(button.isChanged());

    TimeKeeper.setTime(1100);
    button.update();
    TEST_ASSERT_TRUE(button.isActive());
    TEST_ASSERT_FALSE(button.isChanged());

    InputKeeper.setDigitalPinState(0, false);
    button.update();
    TEST_ASSERT_FALSE(button.isActive());
    TEST_ASSERT_TRUE(button.isChanged());
 
}



void setup() {
    UNITY_BEGIN(); 
    RUN_TEST(test_DigitalInput_NoDelay);
    RUN_TEST(test_digitalInput_WithDelay);
    UNITY_END(); 
}

void loop() {
}