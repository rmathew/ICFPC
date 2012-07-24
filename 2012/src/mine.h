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

#define DIFF(A,B) (((A) > (B)) ? ((A) - (B)) : ((B) - (A)))
#define DIST(A,B) (DIFF((A).x, (B).x) + DIFF((A).y, (B).y))

typedef enum {
  PLAYING,
  WON,
  LOST,
  ABORTED,
} status_t;

typedef struct {
  uint16_t x, y;
} pos_t;

#define SAME_POS(A,B) (((A).x == (B).x) && ((A).y == (B).y))

extern int mine_init(FILE* map_fp);
extern bool are_rocks_pinned(void);
extern uint16_t get_num_rows(void);
extern uint16_t get_num_cols(void);
extern char get_entity_at(const pos_t* pos);
extern pos_t* get_orig_lambdas(uint16_t* num_p);
extern int32_t get_score(void);
extern uint16_t get_num_lambdas_left(void);
extern void get_lift_pos(pos_t* pos);
extern void get_robot_pos(pos_t* pos);
extern status_t get_status(void);
extern void refresh_mine(char cmd);
extern const char* get_cmds(void);
extern int mine_quit(void);

#endif /* MINE_H_INCLUDED */
