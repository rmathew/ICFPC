#ifndef MINE_H_INCLUDED
#define MINE_H_INCLUDED

#define ENTITY_LAMBDA '\\'
#define ENTITY_ROBOT 'R'
#define ENTITY_WALL '#'
#define ENTITY_ROCK '*'
#define ENTITY_CLOSED_LIFT 'L'
#define ENTITY_OPEN_LIFT 'O'
#define ENTITY_EARTH '.'
#define ENTITY_EMPTY ' '
#define ENTITY_UNKNOWN '?'

#define CMD_LEFT 'L'
#define CMD_RIGHT 'R'
#define CMD_UP 'U'
#define CMD_DOWN 'D'
#define CMD_WAIT 'W'
#define CMD_ABORT 'A'
#define CMD_UNKNOWN '?'

#define MAX_CMDS 4096U

typedef enum {
  PLAYING,
  WON,
  LOST,
  ABORTED,
} status_t;

extern int mine_init(FILE* map_fp);
extern uint16_t get_num_rows(void);
extern uint16_t get_num_cols(void);
extern char get_entity_at(uint16_t x, uint16_t y);
extern int32_t get_score(void);
extern uint16_t get_num_lambdas_left(void);
extern void get_lift_pos(uint16_t* x, uint16_t* y);
extern void get_robot_pos(uint16_t* x, uint16_t* y);
extern status_t get_status(void);
extern void refresh_mine(char cmd);
extern const char* get_cmds(void);
extern int mine_quit(void);

#endif /* MINE_H_INCLUDED */
