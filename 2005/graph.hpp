#ifndef _GRAPH_INCLUDED_
#define _GRAPH_INCLUDED_

#include <limits>
#include <string>
#include <map>

using namespace std;

const unsigned int infinite_cost = numeric_limits<unsigned int>::max ();

/**
 * Represents a street (edge) in the world map. Every street
 * is one-way; two-way streets are represented (by the server)
 * as two opposing one-way streets.
 */
class edge
{
public:
  string from_node;
  string to_node;
  bool car_only;
};

/**
 * Represents a location (vertex) in the world map (graph) and the
 * streets (edges) leading from it to other locations, if any.
 */
class node
{
public:
  string name;
  string node_tag;
  int x, y;
  map<string,edge*> edges;
};

/**
 * Represents the path information from a given node to another node.
 */
class path_info
{
public:
  unsigned int cost;
  string prev_node;
};

/**
 * Represents the world map as told to us by the server.
 */
class graph
{
public:
  // An adjacency list for the graph representing the world map.
  map<string,node*> nodes;

  graph () { }

  /**
   * Gets the node with the given name.
   * Returns NULL if the node was not found.
   */
  node* get_node (const string& name);

  /**
   * Gets the edge between the given nodes.
   * Returns NULL if the edge was not found.
   */
  edge* get_edge (const string& from_node, const string& to_node);

  /**
   * Returns TRUE only if the given nodes have edges running in both
   * directions.
   */
  bool is_two_way (const string& node1, const string& node2)
  {
    return ((get_edge (node1, node2) != NULL)
            && (get_edge (node2, node1) != NULL));
  }

  /**
   * Determines if we can move from the given source node to
   * the given destination node with a given transport mode.
   */
  bool can_move (const string& from, const string& to, bool on_foot = true);

  /**
   * The cost of moving from FROM to TO. Returns
   * infinity if there is no connecting edge.
   */
  unsigned int cost_to_move (const string& from, const string& to, bool on_foot);

  /**
   * Computes shortest paths from the given node to each of the
   * other nodes using Dijkstra's algorithm.
   */
  void shortest_paths (const string& from, bool on_foot,
                       map<string,path_info*>* path_map);

};

#endif /* _GRAPH_INCLUDED_ */
