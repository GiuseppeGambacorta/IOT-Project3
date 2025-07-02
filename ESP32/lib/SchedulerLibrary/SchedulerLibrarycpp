#include "SchedulerLibrary.h"

Scheduler::Scheduler() {}

void Scheduler::init(int basePeriod) {
  this->basePeriod = basePeriod;
  this->nTasks = 0;
  this->timer.setupPeriod(basePeriod);
}

bool Scheduler::addTask(Task* task) {
  if (nTasks < MAX_TASKS) {
    taskList[nTasks] = task;
    nTasks++;
    return true;
  } else {
    return false;
  }
}

void Scheduler::schedule() {
  timer.waitForNextTick();
  for (int i = 0; i < nTasks; i++) {
    if (taskList[i]->isActive() && taskList[i]->updateAndCheckTime(basePeriod)) {
      ServiceLocator::getTimeKeeperInstance().update();
      taskList[i]->tick();
    }
  }
}

Task** Scheduler::getTaskList(){
  Task** retTaskList;
  return retTaskList = this->taskList;
}

int Scheduler::getNumTask(){
  int retNumTask;
  return retNumTask= this->nTasks;
}

Task* Scheduler::getTask(int index){
  Task* retTask;
  return retTask = this->taskList[index];
}