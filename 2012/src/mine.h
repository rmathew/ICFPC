#ifndef MINE_H_INCLUDED
#define MINE_H_INCLUDED

#define MAX_CMDS 2048U

typedef enum {
  MOVE_LEFT,
  MOVE_RIGHT,
  MOVE_UP,
  MOVE_DOWN,
  WAIT,
  ABORT,
  UNKNOWN,
} robot_cmd_t;

typedef enum {
  LAMBDA = 0,
  MINER,
  OPEN_LIFT,
  CLOSED_LIFT,
  ROCK,
  BRICKS,
  EARTH,
  EMPTY_SPACE,
} entity_t;

typedef enum {
  PLAYING,
  WON,
  LOST,
  ABORTED,
} robot_cond_t;

extern int mine_init(const char* map_name, FILE* map_fp);
extern const char* get_mine_name(void);
extern uint16_t get_num_rows(void);
extern uint16_t get_num_cols(void);
extern entity_t get_entity_at(uint16_t x, uint16_t y);
extern int32_t get_score(void);
extern robot_cond_t get_robot_condition(void);
extern void refresh_mine(robot_cmd_t cmd);
extern const char* get_cmds(void);
extern int mine_quit(void);

#endif /* MINE_H_INCLUDED */
