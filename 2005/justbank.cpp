#include <cstdlib>
#include <ctime>
#include <string>
#include <vector>
#include <map>
#include <queue>

#include "graph.hpp"
#include "world.hpp"
#include "robber.hpp"

using namespace std;

struct candi_node
{
  string name;
  int risk;
  unsigned int cost;
  bool to_looted_bank;

  bool operator< (const candi_node& x) const
  {
    if (risk == x.risk)
    {
      if (to_looted_bank)
      {
        if (!x.to_looted_bank)
        {
          return true;
        }
        else
          return (cost > x.cost);
      }
      else
      {
        if (!x.to_looted_bank)
          return (cost > x.cost);
        else
          return false;
      }
    }
    else
      return (risk > x.risk);
  }
};

just_bank::just_bank (const char* aName)
  : robber (aName)
{
  // Initialise the random number generator with a random seed.
  srand (this->rand_seed ());
}

void just_bank::digest_world_map (void)
{
  // Find out the banks and the shortest paths to them from each node.
  for (unsigned int i = 0; i < this->bank_locations.size(); ++i) 
  {
    const string& a_bank = this->bank_locations[i];
    this->world_map.shortest_paths (a_bank, true, &this->paths2banks[a_bank]);
  }
}

void just_bank::digest_world (void)
{
  string next_loc = this->location;

  priority_queue<candi_node> choices;

  unsigned int num_banks = this->bank_locations.size ();
  for (unsigned int i = 0; i < num_banks; i++)
  {
    candi_node a_choice;

    string a_bank = this->bank_locations[i];

    // Are we looting this bank right now?
    if (this->location == a_bank)
      this->looted_banks[a_bank] = true;
    
    // Have we ever looted it?
    map<string,bool>::const_iterator p
      = this->looted_banks.find (a_bank);

    if (p != this->looted_banks.end ())
      a_choice.to_looted_bank = true;
    else
      a_choice.to_looted_bank = false;
   
    // What path must we take to reach the bank quickly?
    map<string,map<string,path_info*> >::const_iterator q
      = this->paths2banks.find (a_bank);

    map<string,path_info*> path = q->second;
    map<string,path_info*>::const_iterator r = path.find (this->location);
    path_info* pi = r->second;

    a_choice.name = pi->prev_node;
    a_choice.cost = pi->cost;
    a_choice.risk = this->risk (a_choice.name);

    choices.push (a_choice);
  }

  // Have we looted everyone?
  if (this->looted_banks.size () == num_banks)
    this->looted_banks.clear ();

  // So where do we go next?
  candi_node next_node = choices.top ();
  next_loc = next_node.name;

  this->move_to (next_loc, this->ptype);
}
