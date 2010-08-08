#ifndef _WORLD_INCLUDED_
#define _WORLD_INCLUDED_

#include <string>
#include <vector>
#include <map>

using namespace std;

/**
 * Information known about another bot during a game.
 * This info becomes available to a bot
 * either for certain when reading world state
 * or via informs from other bots (which might
 * not be trustworthy). A bot also sends this
 * info as an inform to the server.
 */
class bot_info
{
public:
  // The name of the other bot.
  string name;

  // The location of the other bot.
  string location;

  // The type of the bot.
  string ptype;

  // In which world is this info valid?
  unsigned int world;
 
  // With what certainty do we know this info?
  // Negative certainty => certainty of a bot
  // not being present at a given location.
  int certainty;

  // Whether the other player is on foot.
  // (derived from ptype).
  bool on_foot;
};

/**
 * Represents the state of the world as told to us by the server 
 * at a given turn.
 */
class world
{
public:
  // The world number (turn) we are currently on.
  int world_num;

  // The amount looted by the robber(s) so far.
  unsigned int robbed;

  // Names of all bank locations, mapped to their current known value.
  map<string,unsigned int> banks;

  // Evidence: in what worlds was the robber at the current location?
  vector<int> evidence;

  // Are we smellin' anything nearby?
  unsigned int smell_distance;
  
  // Information about other bots that we are 100% confident of since
  // this is what the server told us in the world description.
  vector<bot_info> cops;
  vector<bot_info> robbers;

  // (For cops) the inform that we sent the server.
  vector<bot_info> sent_inform;

  // (For cops) The informs that we received via the server from other cops.
  // Indexed by the name of the informant.
  map<string, vector<bot_info> > recv_informs;

  // (For cops) the plan that we sent the server.
  vector<bot_info> sent_plan;

  // (For cops) The plans that we received via the server from other cops.
  // Indexed by the name of the proposer.
  map<string, vector<bot_info> > recv_plans;

  // (for cops) The vote for all bot's plans that we sent.
  vector<string> sent_vote;

  // Winner of the election (empty if no winner).
  string winner;

  // The chosen move in this world.
  string dest_location;
  string new_ptype;

  // Initialize the world. Note that this is not just a constructor,
  // the bot's current-world gets reinitialized on each turn.
  // Also note, that the current location and ptype are not changed,
  // these are initialized on startup and updated when moves are made.
  void clear();
};

#endif /* _WORLD_INCLUDED_ */
