#include <string>
#include <vector>
#include <map>

#include "graph.hpp"
#include "bot.hpp"
#include "robber.hpp"

using namespace std;

robber::robber (const char* aName)
  : bot (aName) 
{ 
  this->ptype = "robber";
  this->on_foot = true;
}

bool robber::is_cop (void)
{
  return false;
}

void robber::digest_world_map (void)
{
  /* Nothing to do. */
}

void robber::digest_world (void)
{
  // Stay put.
  this->move_to (this->location, this->ptype);
}

// 0 => No risk
// 500 => Cop just another node away.
// 1000 => Cop sitting right there.
int robber::risk (string& dest)
{
  // Maps cop locations to their "on_foot" status.
  map<string,bool> cops;

  int num_cops = this->current_world.cops.size ();
  for (int i = 0; i < num_cops; i++)
  {
    bot_info a_cop = this->current_world.cops[i];
    cops[a_cop.location] = a_cop.on_foot;
  }

  // See if we're running directly into a cop.
  map<string,bool>::const_iterator p = cops.find (dest);
  if (p != cops.end ())
    return 1000;

  // Now see if the node we are going to is just an edge
  // away from a cop who can reach there.
  map<string,bool>::const_iterator q = cops.begin ();
  while (q != cops.end ())
  {
    if (this->world_map.can_move (q->first, dest, q->second))
      return 500;

    q++;
  }

  // No risk as fas as we can see it.
  return 0;
}
