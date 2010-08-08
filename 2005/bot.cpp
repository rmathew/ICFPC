#include <iostream>
#include <string>
#include <cstring>
#include <cstdlib>

#include "graph.hpp"
#include "bot.hpp"

using namespace std;

#define MAX_LINE_SIZE 1024 * 1024  // Let's be generous.

bot::bot (const char* aName)
  : name (aName), state (INITIAL)
{
  this->current_world.clear();
}

int bot::start ()
{
  // Send a registration message to the server.
  cout << "reg: " << this->name << " " << this->ptype << endl << flush;

  // Ready to read in the world skeleton from the server.
  this->state = WAITING_FOR_WORLD;

  // Act on messages from server as they arrive.

  char buffer[MAX_LINE_SIZE + 2];
  bool game_over = false;

  cin.getline (buffer, MAX_LINE_SIZE);
  while (!game_over && !cin.eof ())
  {
    game_over = parse (buffer);

    if (game_over)
      break;

    cin.getline (buffer, BUFSIZ);
  }

  return 0;
}

bool bot::parse (char* aLine)
{
  char* cmd = ::strtok (aLine, " \t");

  switch (this->state)
  {
  case WAITING_FOR_WORLD:
    if (::strcmp (cmd, "wsk\\") == 0)
    {
      this->state = READING_WORLD_SKELETON;
    }
    else
      parse_snafu (cmd);
    break;

  case READING_WORLD_SKELETON:
    if (::strcmp (cmd, "name:") == 0)
    {
      this->name = ::strtok (NULL, " \t");
    }
    else if (::strcmp (cmd, "cop:") == 0)
    {
      string name (::strtok (NULL, " \t"));
      this->cops.push_back (name);
    }
    else if (::strcmp (cmd, "robber:") == 0)
    {
      string name (::strtok (NULL, " \t"));
      this->robbers.push_back (name);
    }
    else if (::strcmp (cmd, "nod\\") == 0)
    {
      this->state = READING_NODES;
    }
    else if (::strcmp (cmd, "edg\\") == 0)
    {
      this->state = READING_EDGES;
    }
    else if (::strcmp (cmd, "wsk/") == 0)
    {
      // Set up the initial location. Do this before digesting
      // the map in case we need to know where we're starting.
      this->location = this->is_cop() ? this->cop_hq_location
                                      : this->robber_init_location;

      this->digest_world_map ();

      this->state = WAITING_FOR_TURN;
    }
    else
      parse_snafu (cmd);
    break;

  case READING_NODES:
    if (::strcmp (cmd, "nod:") == 0)
    {
      node* a_node = new node ();

      a_node->name = ::strtok (NULL, " \t");

      const char* node_tag = ::strtok (NULL, " \t");
      a_node->node_tag = node_tag;

      a_node->x = ::atoi (::strtok (NULL, " \t"));
      a_node->y = ::atoi (::strtok (NULL, " \t"));

      this->world_map.nodes[a_node->name] = a_node;

      if (::strcmp (node_tag, "robber-start") == 0)
      {
        this->robber_init_location = a_node->name;
      }
      else if (::strcmp (node_tag, "hq") == 0)
      {
        this->cop_hq_location = a_node->name;
      }
      else if (::strcmp (node_tag, "bank") == 0)
      {
        this->bank_locations.push_back(a_node->name);
      }
    }
    else if (::strcmp (cmd, "nod/") == 0)
    {
      this->state = READING_WORLD_SKELETON;
    }
    else
      parse_snafu (cmd);
    break;

  case READING_EDGES:
    if (::strcmp (cmd, "edg:") == 0)
    {
      edge* an_edge = new edge ();
      an_edge->from_node = ::strtok (NULL, " \t");
      an_edge->to_node = ::strtok (NULL, " \t");

      const char* edge_type = ::strtok (NULL, " \t");
      if (::strcmp (edge_type, "car") == 0)
        an_edge->car_only = true;
      else
        an_edge->car_only = false;

      node* from_node = this->world_map.get_node (an_edge->from_node);
      from_node->edges[an_edge->to_node] = an_edge;
    }
    else if (::strcmp (cmd, "edg/") == 0)
    {
      this->state = READING_WORLD_SKELETON;
    }
    else
      parse_snafu (cmd);
    break;

  case WAITING_FOR_TURN:
    if (::strcmp (cmd, "wor\\") == 0)
    {
      this->current_world.clear();
      this->state = READING_WORLD_STATE;
    }
    else if (::strcmp (cmd, "game-over") == 0)
    {
      this->state = GAME_OVER;
    }
    else
      parse_snafu (cmd);
    break;

  case READING_WORLD_STATE:
    if (::strcmp (cmd, "wor:") == 0)
    {
      this->current_world.world_num = ::atoi (::strtok (NULL, " \t"));
    }
    else if (::strcmp (cmd, "rbd:") == 0)
    {
      this->current_world.robbed = ::atoi (::strtok (NULL, " \t"));
    }
    else if (::strcmp (cmd, "bv\\") == 0)
    {
      this->state = READING_BANK_VALUES;
    }
    else if (::strcmp (cmd, "ev\\") == 0)
    {
      this->state = READING_EVIDENCES;
    }
    else if (::strcmp (cmd, "smell:") == 0)
    {
      this->current_world.smell_distance = ::atoi (::strtok (NULL, " \t"));
    }
    else if (::strcmp (cmd, "pl\\") == 0)
    {
      this->state = READING_VISIBLE_PLAYERS;
    }
    else if (::strcmp (cmd, "wor/") == 0)
    {
      // We've received a world description.
      if (!this->is_cop()) 
      {
        // Robber.makes a move.
        this->digest_world ();
        this->state = WAITING_FOR_TURN;
      }
      else
      {
        // Cop.informs the server of what it knows
        this->decide_inform(&this->current_world.sent_inform);
        this->send_inform(this->current_world.sent_inform);
        
        // and waits for information from others.
        this->state = WAITING_FOR_INFORMS;
      }
    }
    else
      parse_snafu (cmd);
    break;

  case READING_BANK_VALUES:
    if (::strcmp (cmd, "bv:") == 0)
    {
      string name (::strtok (NULL, " \t"));
      const char* val_str = ::strtok (NULL, " \t");
      this->current_world.banks[name] = ::atoi (val_str);
    }
    else if (::strcmp (cmd, "bv/") == 0)
    {
      this->state = READING_WORLD_STATE;
    }
    else
      parse_snafu (cmd);
    break;

  case READING_EVIDENCES:
    if (::strcmp (cmd, "ev:") == 0)
    {
      string loc(::strtok(NULL, " \t"));  // Ignore location since it is the current node.
      const char* world_num = ::strtok (NULL, " \t");
      this->current_world.evidence.push_back( ::atoi(world_num) );
    }
    else if (::strcmp (cmd, "ev/") == 0)
    {
      this->state = READING_WORLD_STATE;
    }
    else
      parse_snafu (cmd);
    break;

  case READING_VISIBLE_PLAYERS:
    if (::strcmp (cmd, "pl:") == 0)
    {
      bot_info bi;
      bi.name = ::strtok (NULL, " \t");
      bi.location = ::strtok (NULL, " \t");
      bi.ptype = ::strtok (NULL, " \t");
      bi.world = this->current_world.world_num;
      bi.certainty = 100;
      bi.on_foot = (bi.ptype == "cop-car") ? false : true;

      if (bi.ptype == "robber")
      {
        this->current_world.robbers.push_back(bi);
      }
      else 
      {
        this->current_world.cops.push_back(bi);
      }
    }
    else if (::strcmp (cmd, "pl/") == 0)
    {
      this->state = READING_WORLD_STATE;
    }
    else
      parse_snafu (cmd);
    break;

  case WAITING_FOR_INFORMS:
    if (::strcmp (cmd, "from\\") == 0)
    {
      this->state = READING_ALL_INFORMS;
    }
    else
      parse_snafu (cmd);
    break;

  case READING_ALL_INFORMS:
    if (::strcmp (cmd, "from:") == 0)
    {
       this->read_info.clear();
       this->read_from_bot = ::strtok(NULL, " \t");
       this->state = READING_ONE_INFORM;
    }
    else if (::strcmp (cmd, "from/") == 0)
    {
       // All informs have been received.

       // Cop.sends a plan to the server.
       this->decide_plan(&this->current_world.sent_plan);
       this->send_plan(this->current_world.sent_plan);

       // Now wait for other people's plans.
       this->state = WAITING_FOR_PLANS;
    }
    else
      parse_snafu (cmd);
    break;

  case READING_ONE_INFORM:
    if (::strcmp (cmd, "inf\\") == 0)
    {
       // We're being lazy here. Ideally, we should
       // separate this into a separate state: 
       // READING_ONE_INFORM_HEADER or something.
    }
    else if (::strcmp (cmd, "inf:") == 0)
    {
      bot_info bi;
      bi.name = ::strtok (NULL, " \t");
      bi.location = ::strtok (NULL, " \t");
      bi.ptype = ::strtok (NULL, " \t");
      bi.world = ::atoi(::strtok(NULL, " \t"));
      bi.certainty = ::atoi(::strtok(NULL, " \t"));
      bi.on_foot = (bi.ptype == "cop-car") ? false : true;
 
      this->read_info.push_back(bi);
    }
    else if (::strcmp (cmd, "inf/") == 0)
    {
       // Finished reading the inform from a bot, store it
       this->current_world.recv_informs[this->read_from_bot] = this->read_info;

       // and resume reading the rest.
       this->state = READING_ALL_INFORMS;
    }
    else
      parse_snafu (cmd);
    break;

  case WAITING_FOR_PLANS:
    if (::strcmp (cmd, "from\\") == 0)
    {
      this->state = READING_ALL_PLANS;
    }
    else
      parse_snafu (cmd);
    break;

  case READING_ALL_PLANS:
    if (::strcmp (cmd, "from:") == 0)
    {
       this->read_info.clear();
       this->read_from_bot = ::strtok(NULL, " \t");
       this->state = READING_ONE_PLAN;
    }
    else if (::strcmp (cmd, "from/") == 0)
    {
       // All plans have been received.

       // TODO(plakal) Cop.now send a vote.
       this->decide_vote(&this->current_world.sent_vote);
       this->send_vote(this->current_world.sent_vote);

       // Now wait for election results.
       this->state = WAITING_FOR_RESULTS;
    }
    else
      parse_snafu (cmd);
    break;

  case READING_ONE_PLAN:
    if (::strcmp (cmd, "plan\\") == 0)
    {
       // We're being lazy here. Ideally, we should
       // separate this into a separate state: 
       // READING_ONE_INFORM_HEADER or something.
    }
    else if (::strcmp (cmd, "plan:") == 0)
    {
      bot_info bi;
      bi.name = ::strtok (NULL, " \t");
      bi.location = ::strtok (NULL, " \t");
      bi.ptype = ::strtok (NULL, " \t");
      bi.world = ::atoi(::strtok(NULL, " \t"));
      bi.certainty = 0;
      bi.on_foot = (bi.ptype == "cop-car") ? false : true;
 
      this->read_info.push_back(bi);
    }
    else if (::strcmp (cmd, "plan/") == 0)
    {
       // Finished reading the inform from a bot, store it
       this->current_world.recv_plans[this->read_from_bot] = this->read_info;

       // and resume reading the rest.
       this->state = READING_ALL_PLANS;
    }
    else
      parse_snafu (cmd);
    break;

  case WAITING_FOR_RESULTS:
    if ((::strcmp (cmd, "winner:") == 0)
        || (::strcmp (cmd, "nowinner:") == 0)) 
    {
      if (::strcmp (cmd, "winner:") == 0)
      {
        this->current_world.winner = ::strtok(NULL, " \t");
      }

      // FINALLY! Cop makes a move.
      this->digest_world ();
      this->state = WAITING_FOR_TURN;
    }
    else
      parse_snafu (cmd);
    break;

  default:
    break;
  }

  return (this->state == GAME_OVER);
}

void bot::move_to (const string& dest, const string& new_ptype)
{
  this->current_world.dest_location = dest;
  this->current_world.new_ptype = new_ptype;

  this->save_world();

  cout << "mov: " 
       << dest << " " << new_ptype
       << endl << flush;

  this->location = dest;
  this->ptype = new_ptype;
}

void bot::decide_inform(vector<bot_info>* inform_data)
{
  // By default, send no info.
  inform_data->clear();
}

void bot::decide_plan(vector<bot_info>* plan_data)
{
  // By default, send no plan.
  plan_data->clear();
}

void bot::decide_vote(vector<string>* bot_ranking)
{
  bot_ranking->clear();
  for (unsigned int i = 0; i < this->cops.size(); ++i) {
    bot_ranking->push_back(cops[i]);
  }
}

void bot::send_inform(const vector<bot_info>& inform_data)
{
  cout << "inf\\" << endl << flush;
  for (unsigned int i = 0; i < inform_data.size(); ++i) {
    cout << "inf: "
         << inform_data[i].name << " "
         << inform_data[i].location << " "
         << inform_data[i].ptype << " "
         << inform_data[i].world << " "
         << inform_data[i].certainty << endl << flush;
  }
  cout << "inf/" << endl << flush;
}

void bot::send_plan(const vector<bot_info>& plan_data)
{
  cout << "plan\\" << endl << flush;
  for (unsigned int i = 0; i < plan_data.size(); ++i) {
    cout << "inf: "
         << plan_data[i].name << " "
         << plan_data[i].location << " "
         << plan_data[i].ptype << " "
         << plan_data[i].world << endl << flush;
  }
  cout << "plan/" << endl << flush;
}

void bot::send_vote(const vector<string>& bot_ranking)
{
  cout << "vote\\" << endl << flush;
  for (unsigned int i = 0; i < bot_ranking.size(); ++i) {
    cout << "vote: "
         << bot_ranking[i] << endl << flush;
  }
  cout << "vote/" << endl << flush;
}

void bot::save_world() 
{
  this->past_worlds.push_back(this->current_world);
}
