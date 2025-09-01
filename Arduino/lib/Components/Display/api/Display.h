#ifndef __DISPLAY_H__
#define __DISPLAY_H__

#include <LiquidCrystal_I2C.h>

class Display {
private:
    LiquidCrystal_I2C lcd;
    String currentMessage;
    String oldMessage;
    int columns;
    int rows;
    int actualRow;

public:
    Display(int address, int columns, int rows);
    void init();
    void on();
    void off();
    void write( const char*  message);
    void clear();
    void update();
};

#endif