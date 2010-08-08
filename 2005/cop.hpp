#ifndef _COP_INCLUDED_
#define _COP_INCLUDED_

#include <string>

#include "graph.hpp"
#include "bot.hpp"

using namespace std;

/**
 * The base class for all cop-bots.
 */
class cop :public bot
{
public:
  cop (const char* aName);

  virtual ~cop () { /* FIXME */ }

  virtual bool is_cop (void);

  virtual void digest_world_map (void);

  virtual void decide_inform(vector<bot_info>* inform_data);

  virtual void decide_plan(vector<bot_info>* plan_data);

  virtual void decide_vote(vector<string>* bot_ranking);

  virtual void digest_world (void);
};

class random_cop : public cop
{
private:
  map<string,bool> visited_nodes;

public:
  random_cop (const char* aName);

  virtual ~random_cop () { /* FIXME */ }

  virtual void digest_world_map (void);

  virtual void digest_world (void);
};

class bank_patrol: public cop
{
protected:
  // For each bank location, shortest paths to every other node.
  map<string,map<string,path_info*> > paths2banks;

public:
  bank_patrol (const char* aName);

  virtual ~bank_patrol () { /* FIXME */ }

  virtual void digest_world_map (void);

  virtual void digest_world (void);
};

#endif /* _COP_INCLUDED_ */
