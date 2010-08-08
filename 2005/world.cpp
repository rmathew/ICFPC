
#include <string>
#include "world.hpp"

void world::clear() 
{
  this->world_num = -1;

  this->robbed = 0;

  this->banks.clear();

  this->evidence.clear();

  this->smell_distance = 0;

  this->cops.clear();
  this->robbers.clear();

  this->sent_inform.clear(); 
  this->recv_informs.clear();

  this->sent_plan.clear(); 
  this->recv_plans.clear();

  this->sent_vote.clear();

  this->winner.clear();

  this->dest_location.clear();
  this->new_ptype.clear();
}


