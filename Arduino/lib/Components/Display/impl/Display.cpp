#include "../api/Display.h"
#include "ArduinoStandardLibrary.h"

Display::Display(int address, int columns, int rows)
    : lcd(address, columns, rows), currentMessage(""), oldMessage(""), columns(columns), rows(rows) {}

void Display::init() {
    lcd.init();
    lcd.backlight();
    lcd.clear();
}

void Display::on() {
    lcd.display();
    lcd.backlight();
    lcd.setCursor(0, 0);
}

void Display::off() {
    lcd.noDisplay();
    lcd.noBacklight();
    this->clear();
}

void Display::write(const char* message) {
    this->currentMessage = message;
}

void Display::clear() {
    lcd.clear();
    this->currentMessage = "";
}

void Display::update() {

    if (this->currentMessage == this->oldMessage) {
        return;
    }
    lcd.clear();
    lcd.setCursor(0, 0);
    lcd.print(this->currentMessage);
    this->oldMessage = this->currentMessage;
}
