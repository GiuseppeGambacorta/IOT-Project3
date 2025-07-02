#pragma once
class Task {

  protected:
    int myPeriod;
    int timeElapsed;
    bool active;
    
  public:
    virtual void init(int period){
      myPeriod = period;  
      timeElapsed = 0;
      active = true;
    }

    virtual void tick() = 0;

    bool updateAndCheckTime(int basePeriod){
      timeElapsed += basePeriod;
      if (timeElapsed >= myPeriod){
        timeElapsed = 0;
        return true;
      } else {
        return false; 
      }
    }



    bool isActive(){
      bool retActive = this->active;
      return retActive;
    }

    void setActive(bool active){
      this->active = active;
    }

    virtual void reset() = 0;
  
};

