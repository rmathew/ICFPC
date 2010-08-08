#ifndef _ROBBER_INCLUDED_
#define _ROBBER_INCLUDED_

#include <string>
#include <vector>
#include <map>

#include "graph.hpp"
#include "bot.hpp"

using namespace std;

/**
 * Base class for all robber bots. It's a no-op bot that doesn't move
 * anywhere and stays put.
 */
class robber : public bot
{
public:
  robber (const char* aName);

  virtual ~robber () { /* FIXME */ }

  virtual bool is_cop (void);

  virtual void digest_world_map (void);

  virtual void digest_world (void);

  // Gets the risk value of moving to a destination.
  // 0 is no foreseeable risk.
  virtual int risk (string& dest);
};

/**
 * A robber bot that randomly moves across edges without anything else
 * in mind.
 */
class random_robber : public robber
{
private:
  map<string,bool> visited_nodes;

public:
  random_robber (const char* aName);

  virtual ~random_robber () { /* FIXME */ }

  virtual void digest_world_map (void);

  virtual void digest_world (void);
};

/**
 * A robber that heads for the nearest bank all the time.
 */
class just_bank : public robber
{
protected:
  // Looted banks.
  map<string,bool> looted_banks;

  // For each bank location, shortest paths to every other node.
  map<string,map<string,path_info*> > paths2banks;

public:
  just_bank (const char* aName);

  virtual ~just_bank () { /* FIXME */ }

  virtual void digest_world_map (void);

  virtual void digest_world (void);
};

#endif /* _ROBBER_INCLUDED_ */
