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

    int col = 0;
    int row = 0;
    for (unsigned int i = 0; i < this->currentMessage.length(); i++) {
        char c = this->currentMessage[i];
        if (c =='\0'){
            break;
        }
        if (c == '\n') {
            row++;
            col = 0;
            lcd.setCursor(col, row);
        } else {
            lcd.print(c);
            col++;
            if (col >= columns) {
                col = 0;
                row++;
                lcd.setCursor(col, row);
            }
        }
        if (row >= rows){
            break;
        }
    }

    this->oldMessage = this->currentMessage;
}
