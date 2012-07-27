#include <stdint.h>
#include <stdbool.h>
#include <stdio.h>
#include <stdlib.h>

#include "mine.h"
#include "pqueue.h"
#include "astar.h"

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
      if (get_entity_at(orig_lambdas + i) == ENTITY_LAMBDA) {
        int32_t the_dist
            = astar_path(&robot_pos, (orig_lambdas + i), curr_path);
        if ((the_dist >= 0) && (the_dist < dist_closest)) {
          dist_closest = the_dist;
          curr_goal.x = orig_lambdas[i].x;
          curr_goal.y = orig_lambdas[i].y;
        }
      }
    }
  } else {
    get_lift_pos(&curr_goal);
  }
}

static void
get_path_to_goal(void) {
  pos_t robot_pos;
  get_robot_pos(&robot_pos);

  if (SAME_POS(robot_pos, curr_goal)) {
    path_cmds_left = 0U;
    curr_path[0] = '\0';
  } else {
    path_cmds_left = astar_path(&robot_pos, &curr_goal, curr_path);
  }
}

int
main(int argc, char* argv[]) {
  int ret_status = mine_init(stdin);
  if (ret_status != 0) {
    return 1;
  }

  if (!are_rocks_pinned()) {
    fprintf(stderr, "\nERROR: Sorry, I can't deal with falling rocks yet. \n");
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
      } while ((path_cmds_left > 0U) && (get_status() == PLAYING));
    } else {
      // No way to get to our desired goal.
      emit_cmd(CMD_ABORT);
    }
  }

  mine_quit();
  return 0;
}
