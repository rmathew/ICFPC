#include <string>
#include <map>

#include "graph.hpp"
#include "bot.hpp"

using namespace std;

node* graph::get_node (const string& name)
{
  node* ret_val = NULL;

  map<string,node*>::const_iterator p = this->nodes.find (name);

  if (p != this->nodes.end ())
    ret_val = p->second;

  return ret_val;
}

edge* graph::get_edge (const string& from_node, const string& to_node)
{
  edge* ret_val = NULL;

  node* from = this->get_node (from_node);

  if (from != NULL)
  {
    map<string,edge*>::const_iterator p = from->edges.find (to_node);

    if (p != from->edges.end ())
      ret_val = p->second;
  }

  return ret_val;
}

bool graph::can_move (const string& from, const string& to, bool on_foot)
{
  bool ret_val = false;

  edge* f2t = this->get_edge (from, to);

  if (f2t != NULL)
  {
    ret_val = ((f2t->car_only && on_foot) ? false : true);
  }
  else
  {
    edge* t2f = this->get_edge (to, from);
    
    if (t2f != NULL && on_foot)
    {
      ret_val = !t2f->car_only;
    }
  }

  return ret_val;
}

/**
 * The cost of moving from FROM to TO. Returns
 * infinity if there is no connecting edge.
 */
unsigned int graph::cost_to_move (const string& from, const string& to, bool on_foot)
{
  unsigned int ret_val = infinite_cost;

  if (this->can_move (from, to, on_foot))
  {
    // Do we want a finer-grained cost model?
    ret_val = 1U;
  }

  return ret_val;
}

void graph::shortest_paths (const string& from, bool on_foot,
                            map<string,path_info*>* path_map)
{
  path_map->clear();

  // Special case for getting from FROM to FROM.
  path_info* this_guys_pi = new path_info ();
  this_guys_pi->cost = 0U;
  this_guys_pi->prev_node = from;

  (*path_map)[from] = this_guys_pi;

  // The rest of the nodes.
  map<string,path_info*> v;

  // Initialise costs for moving from FROM to each of the other nodes.
  map<string,node*>::const_iterator p = this->nodes.begin ();
  while (p != this->nodes.end ())
  {
    if (p->first != from)
    {
      path_info* pi = new path_info ();

      pi->cost = cost_to_move (from, p->first, on_foot);
      pi->prev_node = from;   // Bogus?
     
      v[p->first] = pi;
    }
    p++;
  }

  // Now keep picking up the node with the minimum cost to move to
  // and update the costs of the rest with respect to this node, if
  // necessary.

  int num_rest = v.size ();

  for (int i = 0; i < num_rest; i++)
  {
    string min_guy;
    unsigned int min_cost = infinite_cost;
    path_info* min_guys_pi = NULL;

    // Find the node with the minimum total cost.

    map<string,path_info*>::const_iterator p = v.begin ();
    while (p != v.end ())
    {
      if (p->second->cost <= min_cost)
      {
        min_cost = p->second->cost;
        min_guy = p->first;
        min_guys_pi = p->second;
      }
      ++p;
    }

    // Remove this node and keep track of it.
    (*path_map)[min_guy] = min_guys_pi;

    v.erase (min_guy);
    
    // Update the costs for the rest of the nodes.
    map<string,path_info*>::const_iterator q = v.begin ();
    while (q != v.end ())
    {
      unsigned int c2m = cost_to_move (min_guy, q->first, on_foot);

      if (c2m < infinite_cost && min_cost < infinite_cost)
      {
        unsigned int cost_via = min_cost + c2m;
        if (cost_via < q->second->cost)
        {
          q->second->cost = cost_via;
          q->second->prev_node = min_guy;
        }
      }
      ++q;
    }
  }
}
