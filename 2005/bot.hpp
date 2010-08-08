#ifndef _BOT_INCLUDED_
#define _BOT_INCLUDED_

#include <sys/time.h>

#include <iostream>
#include <string>
#include <vector>
#include <map>
#include <cstdlib>

#include "graph.hpp"
#include "world.hpp"

using namespace std;

/**
 * The current state of a bot. This is mainly used to ease parsing messages
 * (see bot.parse()) and can also be used to validate the messages that
 * can be understood in any given state. Refer to the state machines for
 * the bots as given in the problem specification.
 */
enum bot_state
{
  INITIAL = 0,
  WAITING_FOR_WORLD,
  READING_WORLD_SKELETON,
  READING_NODES,
  READING_EDGES,
  WAITING_FOR_TURN,
  READING_WORLD_STATE,
  READING_BANK_VALUES,
  READING_EVIDENCES,
  READING_VISIBLE_PLAYERS,
  WAITING_FOR_INFORMS,
  READING_ALL_INFORMS,
  READING_ONE_INFORM,
  WAITING_FOR_PLANS,
  READING_ALL_PLANS,
  READING_ONE_PLAN,
  WAITING_FOR_RESULTS,
  READING_FROM_MSGS,
  GAME_OVER
};

/**
 * The base class that represents all bots, whether cops or robbers.
 * Currently parses all messages and calls hooks on child classes
 * at various points.
 */
class bot
{
private:
  /**
   * Parses a single message sent by the game server. Uses the state of
   * the bot to determine what messages should be received and automatically
   * advances the state, if neccessary. Calls various hooks on child classes
   * when needed.
   */
  bool parse (char* aLine);

  /**
   * Used to indicate that the bot has received an illegal or unknown
   * command from the server as determined by its state. This should not
   * normally happen, but...
   */
  void parse_snafu (const char* cmd)
    {
      cerr << "ERROR: Unknown command \"" << cmd << "\" in state "
        << this->state << endl;
      
      ::exit (1);
    }

  // Sometimes we receive messages where repeated elements span multiple
  // lines (informs, plans etc). For these cases, we need to keep track
  // of the info being received as one element as well as the key to index 
  // each of these elements. Usually, it's the name of the bot from whom 
  // we're receiving the inform/plan/etc.
  vector<bot_info> read_info;
  string read_from_bot;

protected:
  // The name of the bot (either as given by the player or as assigned by
  // the server).
  string name;

  // Whether the bot is on a car or on foot.
  bool on_foot;

  // All robbers and cops known to this bot (via the world skeleton message).
  vector<string> robbers;
  vector<string> cops;

  // The world map as seen by this bot.
  graph world_map;

  // Names of some fixed locations.
  string cop_hq_location;
  string robber_init_location;
  vector<string> bank_locations;

  // The current state of the internal state machine.
  bot_state state;

  // The name of the current location.
  string location;

  // Our current ptype. Initialized by subclasses. Can be switched by cops.
  string ptype;

  // The world as seen after the latest turn.
  world current_world;

  // An array of all past worlds. This is maintained by
  // the save_world() method and may not be needed by some bots.
  vector<world> past_worlds;

public:
  bot (const char* aName);

  virtual ~bot () { /* FIXME */ }

  /**
   * A random number generator seed generator.
   */
  unsigned int rand_seed (void)
  {
    struct timeval tv;
    struct timezone tz;
    ::gettimeofday (&tv, &tz);
    return (unsigned int)( tv.tv_usec);
  }

  /**
   * Starts communicating with the server by first registering with
   * it and then interpreting its actions (via parse ()). Finishes when
   * the server indicates "game-over".
   */
  int start (void);

  /**
   * Is the current bot a cop or a robber?
   */
  virtual bool is_cop (void) = 0;

  /**
   * Is the current bot travelling on foot?
   */
  bool is_on_foot (void) { return this->on_foot; };

  /**
   * A hook that is called just after the world skeleton has been
   * read in, to enable a bot to digest the world information and to
   * pre-compute stuff, if needed.
   */
  virtual void digest_world_map (void) = 0;

  // For cops: ponder over the current world and decide what
  // inform data we are willing to send to other cops.
  // Default implementation sends no info.
  virtual void decide_inform(vector<bot_info>* inform_data);

  // For cops: ponder over the current world and decide what
  // plan are willing to send to other cops.
  // Default implementation sends no plan.
  virtual void decide_plan(vector<bot_info>* plan_data);

  // For cops: ponder over the current world and decide the
  // order of plans that we want to vote for.
  // Default implementation: list all cops in the order
  // we received in the initial world skeleton.
  virtual void decide_vote(vector<string>* bot_ranking);

  /**
   * Ponder over the current state of the world and make a move.
   */
  virtual void digest_world (void) = 0;

  /**
   * Actually move to the given location by updating internal state
   * and sending a message to the server.
   */
  void move_to (const string& dest, const string& new_ptype);

  // Send a inform message to the server about what we know
  // about other players.
  void send_inform(const vector<bot_info>& inform_data);

  // Send a plan message to the server about what we know
  // about other players.
  void send_plan(const vector<bot_info>& plan_data);

  // Send a vote ordering the plans of all bots.
  void send_vote(const vector<string>& bot_ranking);

  // Save the current world in our history of past worlds.
  // The default implementation saves all worlds. Some bots
  // can provide an empty implementation if they don't care
  // about history.
  virtual void save_world();
};

#endif /* _BOT_INCLUDED_ */
