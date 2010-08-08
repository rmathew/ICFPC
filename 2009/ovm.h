#ifndef OVM_H_INCLUDED
#define OVM_H_INCLUDED

/* Team Identifier for "rmathew". */
#define TEAM_ID 632

#define EARTH_MASS 6.0e+24L

#define EARTH_RADIUS 6.357e+6L

#define GRAVITY_CONST 6.67428e-11L

#define MU (GRAVITY_CONST * EARTH_MASS)

#define DIST_EPSILON 1000.0L

#define MAX_OTH_SATS 11

#define SCENARIO_PORT 0x3E80


extern void load_obf (const char *file);

extern void init_scenario (uint32_t s_id);
extern void quit_scenario (void);

extern void set_inputs (uint32_t n, ...);
extern void run_ovm (void);
extern double get_output (uint32_t p_num);

extern uint32_t get_time_step (void);
extern double get_score (void);
extern double get_fuel (void);

#endif /* OVM_H_INCLUDED */
