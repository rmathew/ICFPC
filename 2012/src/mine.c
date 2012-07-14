#include <stdint.h>
#include <stdbool.h>
#include <string.h>
#include <stdio.h>

#include "mine.h"

#define MAX(a,b) ((a) > (b) ? (a) : (b))
#define MIN(a,b) ((a) < (b) ? (a) : (b))

#define MAX_MAP_SIZE 2048U

static char mine_map[MAX_MAP_SIZE][MAX_MAP_SIZE];

static uint16_t num_rows;
static uint16_t num_cols;

static robot_cond_t robot_condition = PLAYING;

static int32_t robot_x;
static int32_t robot_y;

static uint16_t lambdas_left = 0U;
static uint16_t lambdas_mined = 0U;

static int32_t score;

int
mine_init(int argc, char* argv[]) {
  int ret_status = 0;
  if (argc != 2) {
    fprintf(stderr, "Map file not specified!\n");
    ret_status = 1;
  }

  if (ret_status == 0) {
    FILE* fp = fopen(argv[1], "r");
    if (fp != NULL) {
      num_rows = 0U;
      num_cols = 0U;
      while (fgets(mine_map[num_rows], MAX_MAP_SIZE, fp) != NULL) {
        size_t row_len = strlen(mine_map[num_rows]);
        if (mine_map[num_rows][row_len - 1] == '\n') {
          mine_map[num_rows][row_len - 1] = '\0';
          row_len--;
        }
        
        // FIXME: Validate input characters.

        // Locate the Robot.
        for (int i = 0; i < row_len; i++) {
          if (mine_map[num_rows][i] == 'R') {
            robot_x = i;
            robot_y = num_rows;
            break;
          } else if (mine_map[num_rows][i] == '\\') {
            lambdas_left++;
          }
        }

        num_cols = MAX(num_cols, row_len);
        num_rows++;
      }
      fclose(fp);

      // Pad empty spaces to rows shorter than the longest row.
      for (int i = 0; i < num_rows; i++) {
        char* the_row = mine_map[i];
        size_t row_len = strlen(the_row);
        if (row_len < num_cols) {
          for (int j = row_len; j < num_cols; j++) {
            the_row[j] = ' ';
          }
          the_row[num_cols] = '\0';
        }
      }

      score = 0;
      robot_condition = PLAYING;
    } else {
      fprintf(stderr, "Could not open map file \"%s\"!\n", argv[1]);
      ret_status = 1;
    }
  }

  return ret_status;
}

int
mine_quit(void) {
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

entity_t
get_entity_at(uint16_t row, uint16_t col) {
  entity_t ret_val = EMPTY_SPACE;

  // FIXME: Boundary-checks.
  char e = mine_map[row][col];
  switch (e) {
  case '\\':
    ret_val = LAMBDA;
    break;
  case 'R':
    ret_val = MINER;
    break;
  case 'O':
    ret_val = OPEN_LIFT;
    break;
  case 'L':
    ret_val = CLOSED_LIFT;
    break;
  case '*':
    ret_val = ROCK;
    break;
  case '#':
    ret_val = BRICKS;
    break;
  case '.':
    ret_val = EARTH;
    break;
  case ' ':
    ret_val = EMPTY_SPACE;
    break;
  default:
    break;
  }
  return ret_val;
}

int
get_score(void) {
  return score;
}

void
update_map(robot_cmd_t cmd) {
  int32_t new_x = robot_x;
  int32_t new_y = robot_y;

  switch (cmd) {
  case MOVE_UP:
    new_y = robot_y - 1;
    score--;
    break;
  case MOVE_DOWN:
    new_y = robot_y + 1;
    score--;
    break;
  case MOVE_LEFT:
    new_x = robot_x - 1;
    score--;
    break;
  case MOVE_RIGHT:
    new_x = robot_x + 1;
    score--;
    break;
  case ABORT:
    score += (25 * lambdas_mined);
    robot_condition = ABORTED;
    break;
  case WAIT:
    score--;
  default:
    break;
  }

  if ((new_x != robot_x) || (new_y != robot_y)) {
    bool move_allowed = false;
    if ((new_x >= 0) && (new_x < num_cols) && (new_y >= 0)
        && (new_y < num_rows)) {
      char at_tgt = mine_map[new_y][new_x];
      switch (at_tgt) {
      case ' ':
      case '.':
      case '\\':
      case 'O':
        move_allowed = true;
        break;
      case '*':
        if ((cmd == MOVE_LEFT) && ((new_x - 1) >= 0)
            && (mine_map[new_y][new_x - 1] == ' ')) {
          mine_map[new_y][new_x - 1] = '*';
          move_allowed = true;
        }
        if ((cmd == MOVE_RIGHT) && ((new_x + 1) < num_cols)
            && (mine_map[new_y][new_x + 1] == ' ')) {
          mine_map[new_y][new_x + 1] = '*';
          move_allowed = true;
        }
        break;
      }
    }

    if (move_allowed) {
      if (mine_map[new_y][new_x] == '\\') {
        lambdas_left--;
        lambdas_mined++;
        score += 25;
      } else if (mine_map[new_y][new_x] == 'O') {
        score += (50 * lambdas_mined);
        robot_condition = WON;
      }
      mine_map[robot_y][robot_x] = ' ';
      robot_x = new_x;
      robot_y = new_y;
      mine_map[robot_y][robot_x] = 'R';
    } else {
      cmd = WAIT;
    }
  }

  for (int32_t i = num_rows - 1; i >= 0; i--) {
    for (int32_t j = 0; j < num_cols; j++) {
      int32_t y_below = i + 1;
      char ent = mine_map[i][j];
      switch (ent) {
      case '*':
        if (y_below < num_rows) {
          if (mine_map[y_below][j] == ' ') {
            mine_map[i][j] = ' ';
            mine_map[y_below][j] = '*';
          } else if (mine_map[y_below][j] == '*') {
            int32_t x_right = j + 1;
            int32_t x_left = j - 1;
            if (x_right < num_cols) {
              if ((mine_map[i][x_right] == ' ')
                  && (mine_map[y_below][x_right] == ' ')) {
                mine_map[i][j] = ' ';
                mine_map[y_below][x_right] = '*';
              } else if (x_left >= 0) {
                if ((mine_map[i][x_left] == ' ')
                    && (mine_map[y_below][x_left] == ' ')) {
                  mine_map[i][j] = ' ';
                  mine_map[y_below][x_left] = '*';
                }
              }
            }
          } else if (mine_map[y_below][j] == '\\') {
            int32_t x_right = j + 1;
            if (x_right < num_cols) {
              if ((mine_map[i][x_right] == ' ')
                  && (mine_map[y_below][x_right] == ' ')) {
                mine_map[i][j] = ' ';
                mine_map[y_below][x_right] = '*';
              }
            }
          }
        }
        break;
      case 'L':
        if (lambdas_left == 0) {
          mine_map[i][j] = 'O';
        }
        break;
      }
    }
  }
}

robot_cond_t
get_robot_condition(void) {
  return robot_condition;
}
