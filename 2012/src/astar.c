#include <stdint.h>
#include <stdbool.h>
#include <stdlib.h>
#include <stdio.h>

#include "mine.h"
#include "pqueue.h"
#include "astar.h"

/*
 * Implements the A-Star algorithm as described in the corresponding Wikipedia
 * article as of 15-July-2012:
 * http://en.wikipedia.org/w/index.php?title=A*_search_algorithm&oldid=502398507
 * (http://en.wikipedia.org/wiki/A*_search_algorithm)
 *
 * We use the taxicab metric:
 *
 *   http://en.wikipedia.org/wiki/Taxicab_geometry
 *
 * for the heuristic h(x) in A-Star, which makes it easy to implement it
 * efficiently.
 */

#define MAX_ASTAR_NODES 8192U

typedef struct astar_node {
  pos_t pos;
  uint16_t g_score;
  uint16_t f_score;
  struct astar_node* came_from;
  char reach_cmd;
} astar_node_t;

static astar_node_t all_nodes[MAX_ASTAR_NODES];
static uint16_t num_astar_nodes = 0U;

static const astar_node_t* closed_set[MAX_ASTAR_NODES];
static uint16_t num_elts_closed = 0U;

static int
cmp_astar_nodes(const void* a1, const void* a2) {
  uint16_t f1 = ((astar_node_t*)a1)->f_score;
  uint16_t f2 = ((astar_node_t*)a2)->f_score;
  if (f1 < f2) {
    return -1;
  } else {
    return (f1 - f2);
  }
}

static bool
equal_astar_nodes(const void* a1, const void* a2) {
  pos_t p1 = ((astar_node_t*)a1)->pos;
  pos_t p2 = ((astar_node_t*)a2)->pos;
  return ((p1.x == p2.x) && (p1.y == p2.y));
}

static astar_node_t*
new_astar_node(const pos_t* p) {
  astar_node_t* ret_val = NULL;
  if (num_astar_nodes < MAX_ASTAR_NODES) {
    ret_val = (all_nodes + num_astar_nodes);
    ret_val->pos.x = p->x;
    ret_val->pos.y = p->y;
    ret_val->g_score = UINT16_MAX;
    ret_val->f_score = UINT16_MAX;
    ret_val->came_from = NULL;
    ret_val->reach_cmd = CMD_UNKNOWN;
    num_astar_nodes++;
  } else {
    fprintf(stderr, "ERROR: Cannot create any more nodes.\n");
    exit(1);
  }
  return ret_val;
}

static void
maybe_add_neighbor(uint16_t x, uint16_t y, char in_cmd,
    astar_node_t* neighbors[], uint16_t* found_p) {
  pos_t p;
  p.x = x;
  p.y = y;
  bool valid_neighbor = false;
  char e = get_entity_at(&p);
  switch (e) {
  case ENTITY_LAMBDA:
  case ENTITY_EMPTY:
  case ENTITY_EARTH:
  case ENTITY_OPEN_LIFT:
    valid_neighbor = true;
    break;
  default:
    valid_neighbor = false;
    break;
  }
  if (valid_neighbor) {
    astar_node_t* n = new_astar_node(&p);
    n->reach_cmd = in_cmd;
    neighbors[*found_p] = n;
    *found_p = *found_p + 1U;
  }
}

static int
find_neighbors(const astar_node_t* s, astar_node_t* neighbors[]) {
  uint16_t found = 0U;

  uint16_t s_x = s->pos.x;
  uint16_t s_y = s->pos.y;

  maybe_add_neighbor(s_x + 1U, s_y, CMD_RIGHT, neighbors, &found);
  if (s_x > 0U) {
    maybe_add_neighbor(s_x - 1U, s_y, CMD_LEFT, neighbors, &found);
  }

  maybe_add_neighbor(s_x, s_y + 1U, CMD_DOWN, neighbors, &found);
  if (s_y > 0U) {
    maybe_add_neighbor(s_x, s_y - 1U, CMD_UP, neighbors, &found);
  }

  return found;
}

static bool
have_node_in_closed_set(const astar_node_t* n) {
  bool ret_val = false;
  for (uint16_t i = 0U; i < num_elts_closed; i++) {
    if (equal_astar_nodes(closed_set[i], n)) {
      ret_val = true;
      break;
    }
  }
  return ret_val;
}

static void
add_to_closed_set(const astar_node_t* n) {
  if (!have_node_in_closed_set(n)) {
    closed_set[num_elts_closed++] = n;
  }
}

static void
clear_closed_set(void) {
  num_elts_closed = 0U;
}

static uint16_t
heuristic_cost_estimate(const astar_node_t* a, const astar_node_t* b) {
  return DIST(a->pos, b->pos);
}

int32_t
astar_path(const pos_t* start_pos, const pos_t* goal_pos, char* path) {
  // Reset all the nodes.
  num_astar_nodes = 0U;

  astar_node_t* start = new_astar_node(start_pos);
  start->reach_cmd = CMD_WAIT;

  astar_node_t* goal = new_astar_node(goal_pos);
  goal->reach_cmd = CMD_ABORT;

  start->g_score = 0U;
  start->f_score = heuristic_cost_estimate(start, goal);

  clear_closed_set();
  pqueue_t* open_set = pq_create(MAX_ASTAR_NODES, cmp_astar_nodes);
  pq_insert(open_set, start);

  astar_node_t* path_end_pt = NULL;
  while (!pq_is_empty(open_set)) {
    astar_node_t* current = (astar_node_t*)pq_delmin(open_set);
    if (equal_astar_nodes(current, goal)) {
      path_end_pt = current;
      break;
    }

    add_to_closed_set(current);

    astar_node_t* neighbors[4];
    uint16_t num_neighbors = find_neighbors(current, neighbors);
    for (int i = 0; i < num_neighbors; i++) {
      if (have_node_in_closed_set(neighbors[i])) {
        continue;
      }

      uint16_t tentative_g_score = current->g_score
          + DIST(current->pos, neighbors[i]->pos);
      if (!pq_has_elt(open_set, neighbors[i], equal_astar_nodes)
          || (tentative_g_score < neighbors[i]->g_score)) {
        pq_insert(open_set, neighbors[i]);
        neighbors[i]->came_from = current;
        neighbors[i]->g_score = tentative_g_score;
        neighbors[i]->f_score
            = tentative_g_score + heuristic_cost_estimate(neighbors[i], goal);
      }
    }
  }

  int32_t ret_val = 0;
  if (path_end_pt != NULL) {
    do {
      path[ret_val++] = path_end_pt->reach_cmd;
      path_end_pt = path_end_pt->came_from;
    } while ((path_end_pt != NULL) && !(equal_astar_nodes(path_end_pt, start)));

    // The path obtained so far is the other way around.
    for (int i = 0, j = (ret_val - 1); i < j; i++, j--) {
      char tmp = path[i];
      path[i] = path[j];
      path[j] = tmp;
    }
    path[ret_val] = '\0';
  } else {
    path[0] = '\0';
    ret_val = -1;
  }

  pq_destroy(open_set);

  return ret_val;
}
