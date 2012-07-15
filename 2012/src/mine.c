#include <stdint.h>
#include <stdbool.h>
#include <string.h>
#include <stdio.h>
#include <stdlib.h>

#include "mine.h"

#define MAX(a,b) ((a) > (b) ? (a) : (b))
#define MIN(a,b) ((a) < (b) ? (a) : (b))

#define MAX_MAP_SIZE 2047U

static char cmds[MAX_CMDS];
static uint32_t num_cmds = 0U;

static char mine_map[MAX_MAP_SIZE][MAX_MAP_SIZE];

static uint16_t num_rows;
static uint16_t num_cols;

static status_t status = PLAYING;

#define INVALID_X (UINT16_MAX - 1U)
#define INVALID_Y (UINT16_MAX - 1U)

static uint16_t robot_x = INVALID_X;
static uint16_t robot_y = INVALID_Y;

static uint16_t lift_x = INVALID_X;
static uint16_t lift_y = INVALID_Y;

static uint16_t num_rocks = 0U;

static uint16_t lambdas_left = 0U;
static uint16_t lambdas_mined = 0U;

static int32_t score;

typedef struct {
  int32_t old_x, old_y;
  int32_t new_x, new_y;
} move_t;

static move_t* pending_moves = NULL;
static uint32_t num_pending_moves = 0U;

int
mine_init(FILE* map_fp) {
  num_rows = 0U;
  num_cols = 0U;
  while (fgets(mine_map[num_rows], MAX_MAP_SIZE, map_fp) != NULL) {
    size_t row_len = strlen(mine_map[num_rows]);
    if (mine_map[num_rows][row_len - 1] == '\n') {
      mine_map[num_rows][row_len - 1] = '\0';
      row_len--;
    }
    
    // Locate the robot and lift; count lambdas and rocks.
    for (int i = 0; i < row_len; i++) {
      switch (mine_map[num_rows][i]) {
      case ENTITY_ROBOT:
        if ((robot_x == INVALID_X) && (robot_y == INVALID_Y)) {
          robot_x = i;
          robot_y = num_rows;
        } else {
          fprintf(stderr, "ERROR: Invalid map - multiple robots found.\n");
          return 1;
        }
        break;
      case ENTITY_CLOSED_LIFT:
        if ((lift_x == INVALID_X) && (lift_y == INVALID_Y)) {
          lift_x = i;
          lift_y = num_rows;
        } else {
          fprintf(stderr, "ERROR: Invalid map - multiple lifts found.\n");
          return 1;
        }
        break;
      case ENTITY_OPEN_LIFT:
        fprintf(stderr, "ERROR: Invalid map - an open lift found!\n");
        return 1;
        break;
      case ENTITY_LAMBDA:
        lambdas_left++;
        break;
      case ENTITY_ROCK:
        num_rocks++;
        break;
      case ENTITY_WALL:
      case ENTITY_EARTH:
      case ENTITY_EMPTY:
        break;
      default:
        fprintf(stderr,
            "ERROR: Invalid map - unknown character \"%c\" found.\n",
            mine_map[num_rows][i]);
        return 1;
        break;
      }
    }

    num_cols = MAX(num_cols, row_len);
    num_rows++;
  }

  if ((robot_x == INVALID_X) || (robot_y == INVALID_Y)) {
    fprintf(stderr, "ERROR: Invalid map - no robot found.\n");
    return 1;
  }

  if ((lift_x == INVALID_X) || (lift_y == INVALID_Y)) {
    fprintf(stderr, "ERROR: Invalid map - no lift found.\n");
    return 1;
  }

  // Pad empty spaces to rows shorter than the longest row.
  for (int i = 0; i < num_rows; i++) {
    char* the_row = mine_map[i];
    size_t row_len = strlen(the_row);
    if (row_len < num_cols) {
      for (int j = row_len; j < num_cols; j++) {
        the_row[j] = ENTITY_EMPTY;
      }
      the_row[num_cols] = '\0';
    }
  }

  pending_moves = malloc(num_rocks * sizeof(move_t));
  num_pending_moves = 0U;
  score = 0;
  status = PLAYING;

  return 0;;
}

int
mine_quit(void) {
  if (pending_moves != NULL) {
    free(pending_moves);
  }
  return 0;
}

uint16_t
get_num_rows(void) {
  return num_rows;
}

uint16_t
get_num_cols(void) {
  return num_cols;
}

char
get_entity_at(uint16_t x, uint16_t y) {
  if ((x < num_cols) && (y < num_rows)) {
    return mine_map[y][x];
  } else {
    return ENTITY_UNKNOWN;
  }
}

int
get_score(void) {
  return score;
}

static void
move_robot(char cmd) {
  int32_t new_x = robot_x;
  int32_t new_y = robot_y;

  switch (cmd) {
  case CMD_UP:
    new_y = robot_y - 1;
    score--;
    break;
  case CMD_DOWN:
    new_y = robot_y + 1;
    score--;
    break;
  case CMD_LEFT:
    new_x = robot_x - 1;
    score--;
    break;
  case CMD_RIGHT:
    new_x = robot_x + 1;
    score--;
    break;
  case CMD_ABORT:
    score += (25 * lambdas_mined);
    status = ABORTED;
    break;
  case CMD_WAIT:
    score--;
  default:
    fprintf(stderr, "ERROR: Unknown command \"%c\".\n", cmd);
    cmd = CMD_ABORT;
    break;
  }

  if (num_cmds < MAX_CMDS) {
    cmds[num_cmds++] = cmd;
  }

  if ((new_x != robot_x) || (new_y != robot_y)) {
    bool move_allowed = false;
    if ((new_x >= 0) && (new_x < num_cols) && (new_y >= 0)
        && (new_y < num_rows)) {
      char at_tgt = mine_map[new_y][new_x];
      switch (at_tgt) {
      case ENTITY_EMPTY:
      case ENTITY_EARTH:
      case ENTITY_LAMBDA:
      case ENTITY_OPEN_LIFT:
        move_allowed = true;
        break;
      case ENTITY_ROCK:
        if ((cmd == CMD_LEFT) && ((new_x - 1) >= 0)
            && (mine_map[new_y][new_x - 1] == ENTITY_EMPTY)) {
          mine_map[new_y][new_x - 1] = ENTITY_ROCK;
          move_allowed = true;
        }
        if ((cmd == CMD_RIGHT) && ((new_x + 1) < num_cols)
            && (mine_map[new_y][new_x + 1] == ENTITY_EMPTY)) {
          mine_map[new_y][new_x + 1] = ENTITY_ROCK;
          move_allowed = true;
        }
        break;
      }
    }

    if (move_allowed) {
      if (mine_map[new_y][new_x] == ENTITY_LAMBDA) {
        lambdas_left--;
        lambdas_mined++;
        score += 25;
      } else if (mine_map[new_y][new_x] == ENTITY_OPEN_LIFT) {
        score += (50 * lambdas_mined);
        status = WON;
      }
      mine_map[robot_y][robot_x] = ENTITY_EMPTY;
      robot_x = new_x;
      robot_y = new_y;
      mine_map[robot_y][robot_x] = ENTITY_ROBOT;
    } else {
      cmd = CMD_WAIT;
    }
  }
}

static void
add_pending_move(int32_t x0, int32_t y0, int32_t x1, int32_t y1) {
  move_t* updt = (pending_moves + num_pending_moves);
  updt->old_x = x0;
  updt->old_y = y0;
  updt->new_x = x1;
  updt->new_y = y1;
  num_pending_moves++;
}

static void
update_map(void) {
  num_pending_moves = 0U;
  for (int32_t i = num_rows - 1; i >= 0; i--) {
    for (int32_t j = 0; j < num_cols; j++) {
      int32_t y_below = i + 1;
      char ent = mine_map[i][j];
      switch (ent) {
      case ENTITY_ROCK:
        if (y_below < num_rows) {
          if (mine_map[y_below][j] == ENTITY_EMPTY) {
            add_pending_move(j, i, j, y_below);
          } else if (mine_map[y_below][j] == ENTITY_ROCK) {
            int32_t x_right = j + 1;
            int32_t x_left = j - 1;
            if (x_right < num_cols) {
              if ((mine_map[i][x_right] == ENTITY_EMPTY)
                  && (mine_map[y_below][x_right] == ENTITY_EMPTY)) {
                add_pending_move(j, i, x_right, y_below);
              } else if (x_left >= 0) {
                if ((mine_map[i][x_left] == ENTITY_EMPTY)
                    && (mine_map[y_below][x_left] == ENTITY_EMPTY)) {
                  add_pending_move(j, i, x_left, y_below);
                }
              }
            }
          } else if (mine_map[y_below][j] == ENTITY_LAMBDA) {
            int32_t x_right = j + 1;
            if (x_right < num_cols) {
              if ((mine_map[i][x_right] == ENTITY_EMPTY)
                  && (mine_map[y_below][x_right] == ENTITY_EMPTY)) {
                add_pending_move(j, i, x_right, y_below);
              }
            }
          }
        }
        break;
      case ENTITY_CLOSED_LIFT:
        if (lambdas_left == 0) {
          mine_map[i][j] = ENTITY_OPEN_LIFT;
        }
        break;
      }
    }
  }

  if (num_pending_moves > 0U) {
    for (int i = 0; i < num_pending_moves; i++) {
      move_t* updt = (pending_moves + i);
      mine_map[updt->old_y][updt->old_x] = ENTITY_EMPTY;
      mine_map[updt->new_y][updt->new_x] = ENTITY_ROCK;
    }
  }
}

void
refresh_mine(char cmd) {
  if (status != PLAYING) {
    return;
  }

  if (cmd == CMD_UNKNOWN) {
    return;
  }

  move_robot(cmd);

  update_map();

  // Check for losing condition if the game has not yet been won or aborted.
  if ((status != WON) && (status != ABORTED)) {
    if ((num_pending_moves > 0U) && ((robot_y - 1) >= 0)
        && (mine_map[robot_y - 1][robot_x] == ENTITY_ROCK)) {
      for (int i = 0; i < num_pending_moves; i++) {
        move_t* updt = (pending_moves + i);
        if ((updt->new_y == (robot_y - 1)) && (updt->new_x == robot_x)) {
          status = LOST;
          break;
        }
      }
    }
  }
}

status_t
get_status(void) {
  return status;
}

const char*
get_cmds(void) {
  cmds[num_cmds] = '\0';
  return cmds;
}

void
get_lift_pos(uint16_t* x, uint16_t* y) {
  *x = lift_x;
  *y = lift_y;
}

void
get_robot_pos(uint16_t* x, uint16_t* y) {
  *x = robot_x;
  *y = robot_y;
}

uint16_t
get_num_lambdas_left(void) {
  return lambdas_left;
}
