#include <stdint.h>
#include <stdbool.h>
#include <stdio.h>
#include <stdlib.h>

#include "mine.h"
#include "pqueue.h"

#define DIFF(A,B) (((A) > (B)) ? ((A) - (B)) : ((B) - (A)))
#define DIST(A,B) (DIFF((A).x, (B).x) + DIFF((A).y, (B).y))

#define MAX_ASTAR_NODES 1024U

typedef struct astar_node {
  pos_t pos;
  uint16_t g_score;
  uint16_t f_score;
  char reach_cmd;
  struct astar_node* came_from;
} astar_node_t;

static pos_t curr_goal;

static uint16_t path_cmds_left;
static char curr_path[4096];

static void
emit_cmd(char cmd) {
  refresh_mine(cmd);
  fputc(cmd, stdout);
}

static void
find_next_goal(void) {
  if (get_num_lambdas_left() > 0U) {
    pos_t robot_pos;
    get_robot_pos(&robot_pos);

    uint16_t dist_closest = UINT16_MAX;
    uint16_t num_orig_lambdas;
    pos_t* orig_lambdas = get_orig_lambdas(&num_orig_lambdas);
    for (int i = 0; i < num_orig_lambdas; i++) {
      uint16_t the_dist = DIST(robot_pos, orig_lambdas[i]);
      if (the_dist < dist_closest) {
        dist_closest = the_dist;
        curr_goal.x = orig_lambdas[i].x;
        curr_goal.y = orig_lambdas[i].y;
      }
    }
  } else {
    get_lift_pos(&curr_goal);
  }
}

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
    astar_node_t* n = (astar_node_t*)malloc(sizeof(astar_node_t));
    n->pos.x = x;
    n->pos.y = y;
    n->g_score = UINT16_MAX;
    n->f_score = UINT16_MAX;
    n->came_from = NULL;
    n->reach_cmd = in_cmd;

    neighbors[*found_p] = n;
    *found_p = *found_p + 1U;
  }
}

static int
find_neighbors(astar_node_t* s, astar_node_t* neighbors[]) {
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
has_node(astar_node_t* set[], uint16_t max, astar_node_t* n) {
  bool ret_val = false;
  for (int i = 0; i < max; i++) {
    if (equal_astar_nodes(set[i], n)) {
      ret_val = true;
      break;
    }
  }
  return ret_val;
}

static uint16_t
astar_path(const pos_t* start_pos, const pos_t* goal_pos, char* path) {
  uint16_t ret_val = 0U;
  astar_node_t* path_end_pt = NULL;

  astar_node_t* start = (astar_node_t*)malloc(sizeof(astar_node_t));
  start->pos.x = start_pos->x;
  start->pos.y = start_pos->y;
  start->came_from = NULL;

  astar_node_t* goal = (astar_node_t*)malloc(sizeof(astar_node_t));
  goal->pos.x = goal_pos->x;
  goal->pos.y = goal_pos->y;
  goal->came_from = NULL;
  goal->g_score = UINT16_MAX;
  goal->f_score = UINT16_MAX;

  start->g_score = 0U;
  start->f_score = DIST(start->pos, goal->pos);

  astar_node_t* closed_set[MAX_ASTAR_NODES];
  uint16_t num_elts_closed = 0U;

  pqueue_t* open_set = pq_create(MAX_ASTAR_NODES, cmp_astar_nodes);
  pq_insert(open_set, start);

  while (!pq_is_empty(open_set)) {
    astar_node_t* current = (astar_node_t*)pq_delmin(open_set);
    if (equal_astar_nodes(current, goal)) {
      path_end_pt = current;
      break;
    }

    closed_set[num_elts_closed++] = current;

    astar_node_t* neighbors[4];
    uint16_t num_neighbors = find_neighbors(current, neighbors);
    for (int i = 0; i < num_neighbors; i++) {
      if (has_node(closed_set, num_elts_closed, neighbors[i])) {
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
            = tentative_g_score + DIST(neighbors[i]->pos, goal->pos);
      }
    }
  }

  if (path_end_pt != NULL) {
    do {
      path[ret_val++] = path_end_pt->reach_cmd;
      path_end_pt = path_end_pt->came_from;
    } while (path_end_pt != NULL);

    // The path is in reverse.
    int i, j;
    char c;
    for (i = 0, j = (ret_val - 1); i < j; i++, j--) {
      c = path[i];
      path[i] = path[j];
      path[j] = c;
    }
    path[ret_val] = '\0';
  }

  pq_destroy(open_set);

  return ret_val;
}

static void
get_path_to_goal(void) {
  pos_t robot_pos;
  get_robot_pos(&robot_pos);
  path_cmds_left = astar_path(&robot_pos, &curr_goal, curr_path);
}

int
main(int argc, char* argv[]) {
  int ret_status = mine_init(stdin);
  if (ret_status != 0) {
    return 1;
  }

  if (!are_rocks_pinned()) {
    fprintf(stderr, "ERROR: Sorry, I can't deal with falling rocks yet. \n");
    return 1;
  }

  while (get_status() == PLAYING) {
    find_next_goal();
    get_path_to_goal();

    if (path_cmds_left > 0U) {
      int cmd_idx = 0;
      do {
        emit_cmd(curr_path[cmd_idx]);
        cmd_idx++;
        path_cmds_left--;
      } while ((path_cmds_left > 0) && (get_status() == PLAYING));
    } else {
      // No way to get to our desired goal.
      emit_cmd(CMD_ABORT);
    }
  }

  mine_quit();
  return 0;
}
