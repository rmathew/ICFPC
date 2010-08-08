#include <cstdlib>
#include <ctime>
#include <vector>
#include "graph.hpp"
#include "cop.hpp"

using namespace std;

random_cop::random_cop (const char* aName)
  : cop (aName)
{
  // Initialise the random number generator with a random seed.
  srand (this->rand_seed ());

  if (rand() < RAND_MAX/2) {
    this->ptype = "cop-foot",
    this->on_foot = true;
  } else {
    this->ptype = "cop-car",
    this->on_foot = false;
  }
}

void random_cop::digest_world_map (void)
{
  map<string,node*>::const_iterator p = this->world_map.nodes.begin ();

  while (p != this->world_map.nodes.end ())
  {
    this->visited_nodes[p->first] = false;
    p++;
  }
  visited_nodes[this->location] = true;
}

void random_cop::digest_world (void)
{
  // At HQ, we can choose to switch transport.
  bool next_on_foot = this->on_foot;
  if ((this->location == this->cop_hq_location) 
      && (rand() < (RAND_MAX/2))) 
  {
    next_on_foot = !next_on_foot;
  }

  vector<string> pref_candis;
  vector<string> all_candis;

  node* curr_node = this->world_map.get_node (this->location);
  map<string,edge*>::const_iterator p = curr_node->edges.begin ();

  while (p != curr_node->edges.end ())
  {
    string tgt_node = p->second->to_node;
    if (this->world_map.can_move (this->location, tgt_node, next_on_foot))
    {
      map<string,bool>::const_iterator q
        = this->visited_nodes.find (tgt_node);

      if (q != this->visited_nodes.end () && !q->second)
        pref_candis.push_back (tgt_node);

      all_candis.push_back (tgt_node);
    }

    p++;
  }

  unsigned int num_pref_candis = pref_candis.size ();
  unsigned int num_all_candis = all_candis.size ();

  unsigned int idx;
  string next_loc;

  if (num_pref_candis > 0)
  {
    idx
      = (int )(((double )(rand ()) / (double) RAND_MAX) * num_pref_candis);
    next_loc = pref_candis[idx];
    this->visited_nodes[next_loc] = true;
  }
  else if (num_all_candis > 0)
  {
    idx
      = (int )(((double )(rand ()) / (double) RAND_MAX) * num_all_candis);
    next_loc = all_candis[idx];
  }
  else
    next_loc = this->location;

  this->move_to (next_loc, next_on_foot ? "cop-foot" : "cop-car");
  this->on_foot = (this->ptype == "cop-foot");
}
