#include <cstdlib>
#include <ctime>
#include <vector>
#include <map>

#include "graph.hpp"
#include "robber.hpp"

using namespace std;

random_robber::random_robber (const char* aName)
  : robber (aName)
{
  // Initialise the random number generator with a random seed.
  srand (this->rand_seed ());
}

void random_robber::digest_world_map (void)
{
  map<string,node*>::const_iterator p = this->world_map.nodes.begin ();

  while (p != this->world_map.nodes.end ())
  {
    this->visited_nodes[p->first] = false;
    p++;
  }
  visited_nodes[this->location] = true;
}

void random_robber::digest_world (void)
{
  vector<string> pref_candis;
  vector<string> all_candis;

  node* curr_node = this->world_map.get_node (this->location);
  map<string,edge*>::const_iterator p = curr_node->edges.begin ();

  while (p != curr_node->edges.end ())
  {
    string tgt_node = p->second->to_node;
    if (this->world_map.can_move (this->location, tgt_node)
        && this->risk (tgt_node) == 0)
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

  this->move_to (next_loc, this->ptype);
}
