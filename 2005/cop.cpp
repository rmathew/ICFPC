#include "cop.hpp"

cop::cop (const char* aName) 
  : bot (aName)
{ 
  this->ptype = "cop-foot",
  this->on_foot = true;
}

bool cop::is_cop (void)
{
  return true;
}

void cop::digest_world_map (void)
{
  /* Nothing to do. */
}

void cop::decide_inform(vector<bot_info>* inform_data)
{
  inform_data->clear();
}

void cop::decide_plan(vector<bot_info>* plan_data)
{
  plan_data->clear();
}

void cop::decide_vote(vector<string>* bot_ranking)
{
  bot_ranking->clear();
  for (unsigned int i = 0; i < this->cops.size(); ++i) {
    bot_ranking->push_back(cops[i]);
  }
}

void cop::digest_world (void)
{
  // Stay put.
  this->move_to (this->location, this->ptype);
  this->on_foot = (this->ptype == "cop-foot");
}
