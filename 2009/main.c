#include <stdint.h>
#include <stdbool.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <math.h>

#include "ovm.h"
#include "ui.h"

typedef struct
{
  long double dv;
  long double dvp;

  long double tv_x_mul;
  long double tv_y_mul;

  uint32_t xfr_time;
} hto_params;

typedef enum
{
  INIT_HTO_IMP,
  FINAL_HTO_IMP,
} hto_imp;


static uint32_t scenario_id;

static char *problem_name;

static char *obf_file;

static void (*problem_solver)(void);

static earth_size es = MEDIUM_EARTH;

static double pending_thrust_x = 0.0;
static double pending_thrust_y = 0.0;


/* Get the angle in radians (0 to 2*PI) for a given point (x,y). */
static long double
pt_angle (long double x, long double y)
{
  /* NOTE: atan () returns values in the range -PI/2 to +PI/2. */
  long double angle = atanl (y/x);

  if (x < 0.0L)
    angle += M_PIl;

  if (angle < 0.0L)
    angle += (2.0L * M_PIl);

  return angle;
}


/* Find if a rotation from angle a1 to a2 is clockwise. */
static bool
is_clockwise (long double a1, long double a2)
{
  long double dif = a2 - a1;

  return (((dif < 0.0L) && (dif > -M_PIl)) || (dif > M_PIl));
}


static void
calc_hto_params (hto_params *ht, long double r1, long double r2, bool cw_orbit)
{
  ht->dv = sqrtl (MU / r1) * (sqrtl (2.0L * r2 / (r1 + r2)) - 1.0L);
  ht->dvp = sqrtl (MU / r2) * (1.0L - sqrtl (2.0L * r1 / (r1 + r2)));

  ht->tv_x_mul = cw_orbit ? -1.0L : +1.0L;
  ht->tv_y_mul = cw_orbit ? +1.0L : -1.0L;

  double r1_r2 = r1 + r2;
  ht->xfr_time = roundl (M_PIl * sqrtl (r1_r2 * r1_r2 * r1_r2 / (8.0L * MU)));
}


static void
give_hto_impulse (hto_params *ht, hto_imp which, long double x, long double y)
{
  long double r = sqrtl (x*x + y*y);

  long double d = (which == INIT_HTO_IMP) ? ht->dv : ht->dvp;

  set_inputs
    (2,
     0x2, (double) (d * y / r * ht->tv_x_mul),
     0x3, (double) (d * x / r * ht->tv_y_mul));
}


/*
static void
catch_up (long double x0, long double y0, long double t_x0, long double t_y0)
{
  // FIXME: Assumes circular and clockwise orbits.

  // Where will the target satellite be in one second?
  // Just gravity: s1 = s0 + v0 + 0.5 * g0
  long double t_r0 = sqrtl (t_x0 * t_x0 + t_y0 * t_y0);

  long double t_g_mag0 = MU / (t_r0 * t_r0);
  long double t_g_x0 = -t_x0 * t_g_mag0 / t_r0;
  long double t_g_y0 = -t_y0 * t_g_mag0 / t_r0;

  long double t_v_mag0 = sqrtl (MU / t_r0);
  long double t_v_x0 = +t_y0 * t_v_mag0 / t_r0;
  long double t_v_y0 = -t_x0 * t_v_mag0 / t_r0;

  long double t_x1 = t_x0 + t_v_x0 + (0.5 * t_g_x0);
  long double t_y1 = t_y0 + t_v_y0 + (0.5 * t_g_y0);


  // Get our satellite there in one second.
  // Gravity and thrusters: s1 = s0 + v0 + 0.5 * (g0 + dV0)
  // Or: dV0 = 2.0 * (s1 - s0 - v0) - g0
  long double r0 = sqrtl (x0 * x0 + y0 * y0);

  long double g_mag0 = MU / (r0 * r0);
  long double g_x0 = -x0 * g_mag0 / r0;
  long double g_y0 = -y0 * g_mag0 / r0;

  long double v_mag0 = sqrtl (MU / r0);
  long double v_x0 = +y0 * v_mag0 / r0;
  long double v_y0 = -x0 * v_mag0 / r0;

  long double dv_x0 = 2.0L * (t_x1 - x0 - v_x0) - g_x0;
  long double dv_y0 = 2.0L * (t_y1 - y0 - v_y0) - g_y0;


  // Fire the thrusters in the *opposite* direction.
  set_inputs (2, 0x2, (double) -dv_x0, 0x3, (double) -dv_y0);

  
  // What will be our velocity then?
  // v1 = v0 + dV0 + 0.5 * (g0 + g1)

  // NOTE: (x1,y1) = (t_x1,t_y1) and r1 = t_r1 = t_r0
  long double g_mag1 = MU / (t_r0 * t_r0);
  long double g_x1 = -t_x1 * g_mag1 / t_r0;
  long double g_y1 = -t_y1 * g_mag1 / t_r0;

  long double v_x1 = v_x0 + dv_x0 + 0.5L * (g_x0 + g_x1);
  long double v_y1 = v_y0 + dv_y0 + 0.5L * (g_y0 + g_y1);


  // Where will the target be after *two* seconds?
  // s2 = s1 + v1 + 0.5 * g1
  long double t_g_x1 = -t_x1 * t_g_mag0 / t_r0;
  long double t_g_y1 = -t_y1 * t_g_mag0 / t_r0;

  long double t_v_x1 = +t_y1 * t_v_mag0 / t_r0;
  long double t_v_y1 = -t_x1 * t_v_mag0 / t_r0;

  long double t_x2 = t_x1 + t_v_x1 + (0.5 * t_g_x1);
  long double t_y2 = t_y1 + t_v_y1 + (0.5 * t_g_y1);

  // What will be its velocity after two seconds?
  // v2 = v1 + 0.5 * (g1 + g2)
  long double t_g_x2 = -t_x2 * t_g_mag0 / t_r0;
  long double t_g_y2 = -t_y2 * t_g_mag0 / t_r0;

  long double t_v_x2 = t_v_x1 + 0.5L * (t_g_x1 + t_g_x2);
  long double t_v_y2 = t_v_y1 + 0.5L * (t_g_y1 + t_g_y2);

 
  // We want to have the same velocity at that point.
  // v2 = v1 + dV1 + 0.5 * (g1 + g2)
  // Or: dV = v2 - v1 - 0.5 * (g1 + g2)
  // So v2 = t_v2. Assume g2 ~= t_g2
  long double dv_x1 = t_v_x2 - v_x1 - 0.5L * (g_x1 + t_g_x2);
  long double dv_y1 = t_v_y2 - v_y1 - 0.5L * (g_y1 + t_g_y2);

  // Remember to thrust in the other direction.
  pending_thrust_x = -dv_x1;
  pending_thrust_y = -dv_y1;
}
*/

/*
static double
get_orbital_period (double a)
{
  return 2 * M_PI * sqrt ((a * a * a) / MU);
}
*/


static void
hohmann_solver (void)
{
  set_inputs (1, SCENARIO_PORT, (double) scenario_id);

  /* Get initial position. */
  run_ovm ();
  long double x0 = get_output (0x2);
  long double y0 = get_output (0x3);

  printf ("\nInitial Fuel: %f\n", get_fuel ());

  /* Get next position to determine the direction of the orbit. */
  run_ovm ();
  long double x1 = get_output (0x2);
  long double y1 = get_output (0x3);

  long double r1 = sqrtl (x1*x1 + y1*y1);
  long double r2 = get_output (0x4);

  bool clockwise = is_clockwise (pt_angle (x0, y0), pt_angle (x1, y1));

  hto_params ht;
  calc_hto_params (&ht, r1, r2, clockwise);

  give_hto_impulse (&ht, INIT_HTO_IMP, x1, y1);

  bool done_dv2 = false;
  uint32_t dv2_tgt_time = get_time_step () + ht.xfr_time;

  bool reset_thrust = true;

  ui_data uid;
  uid.have_fs = false;
  uid.num_oth = 0;

  bool cont_sim = true;

  while (cont_sim)
  {
    run_ovm ();

    uid.x = get_output (0x2);
    uid.y = get_output (0x3);

    cont_sim = update_ui (&uid);

    if (!cont_sim)
    {
      long double r = sqrtl (uid.x * uid.x + uid.y * uid.y);
      printf ("\nOrbiting at %.2Lf m (v/s %.2Lf m planned)\n", r, r2);
    }


    if (reset_thrust)
    {
      set_inputs (2, 0x2, (double) 0.0, 0x3, (double) 0.0);
      reset_thrust = false;
    }
    else if (!done_dv2)
    {
      if (get_time_step () == dv2_tgt_time)
      {
        give_hto_impulse (&ht, FINAL_HTO_IMP, uid.x, uid.y);

        done_dv2 = true;
        reset_thrust = true;
      }
    }
  }
}


static void
mng_solver (void)
{
  set_inputs (1, SCENARIO_PORT, (double) scenario_id);

  /* Get initial position. */
  run_ovm ();
  long double x0 = get_output (0x2);
  long double y0 = get_output (0x3);

  printf ("\nInitial Fuel: %f\n", get_fuel ());

  /* Get next position to determine the direction of the orbit. */
  run_ovm ();
  long double x1 = get_output (0x2);
  long double y1 = get_output (0x3);

  bool clockwise = is_clockwise (pt_angle (x0, y0), pt_angle (x1, y1));

  long double r1 = sqrtl (x1*x1 + y1*y1);

  /* Get target satellite's position and orbit radius. */
  long double x2 = x1 - (long double) get_output (0x4);
  long double y2 = y1 - (long double) get_output (0x5);

  long double r2 = sqrtl (x2*x2 + y2*y2);

  /* Angle in radians swept per second at the moment. */
  long double ang_per_sec = sqrtl (MU / (r1 * r1 * r1));

  hto_params ht;
  calc_hto_params (&ht, r1, r2, clockwise);

  long double tmp = (r1 + r2) / (2.0L * r2);
  long double phi = M_PIl * sqrtl (tmp * tmp * tmp);
  long double tgt_phase_dif = M_PIl - phi;

  bool done_dv1 = false;
  bool done_dv2 = false;
  uint32_t dv2_tgt_time = 0;
  bool reset_thrust = false;
  bool sync_tgt_phase = true;

  ui_data uid;
  uid.have_fs = false;
  uid.num_oth = 1;

  bool cont_sim = true;

  while (cont_sim)
  {
    run_ovm ();

    uid.x = get_output (0x2);
    uid.y = get_output (0x3);

    uid.oth_x[0] = uid.x - (long double) get_output (0x4);
    uid.oth_y[0] = uid.y - (long double) get_output (0x5);

    cont_sim = update_ui (&uid);

    if (!cont_sim)
    {
      long double r = sqrtl (uid.x * uid.x + uid.y * uid.y);
      printf ("\nOrbiting at %.2Lf m (v/s %.2Lf m planned)\n", r, r2);

      long double dx = get_output (0x4);
      long double dy = get_output (0x5);
      r = sqrtl (dx*dx + dy*dy);
      printf ("Target %.2Lf m away\n", r);
    }


    if (reset_thrust)
    {
      set_inputs (2, 0x2, pending_thrust_x, 0x3, pending_thrust_y);

      pending_thrust_x = 0.0;
      pending_thrust_y = 0.0;

      reset_thrust = false;
    }
    else if (!done_dv1)
    {
      /* FIXME: Assumes clockwise rotation of both satellites. */
      /* FIXME: Assumes target is outside. */

      bool do_dv1 = true;

      if (sync_tgt_phase)
      {
        long double trigger_ang
          = tgt_phase_dif + pt_angle (uid.oth_x[0], uid.oth_y[0]);

        if (trigger_ang > 2.0L * M_PIl)
          trigger_ang -= 2.0L * M_PIl;

        long double own_ang = pt_angle (uid.x, uid.y);

        do_dv1
          = (own_ang > trigger_ang)
            && ((own_ang - ang_per_sec) <= trigger_ang);
      }

      if (do_dv1)
      {
        give_hto_impulse (&ht, INIT_HTO_IMP, uid.x, uid.y);
        done_dv1 = true;

        pending_thrust_x = 0.0;
        pending_thrust_y = 0.0;
        reset_thrust = true;

        dv2_tgt_time = get_time_step () + ht.xfr_time;
      }
    }
    else if (!done_dv2)
    {
      if (get_time_step () == dv2_tgt_time)
      {
        give_hto_impulse (&ht, FINAL_HTO_IMP, uid.x, uid.y);

        done_dv2 = true;

        pending_thrust_x = 0.0;
        pending_thrust_y = 0.0;
        reset_thrust = true;
      }
    }
    else
    {
      long double dx = get_output (0x4);
      long double dy = get_output (0x5);
      long double tgt_dist = sqrtl (dx*dx + dy*dy);

      if (tgt_dist >= DIST_EPSILON)
      {
        // FIXME: Clockwise circular orbit assumed.
        // XXX: What to do here?
      }
    }
  }
}


static void
ecc_mng_solver (void)
{
  set_inputs (1, SCENARIO_PORT, (double) scenario_id);

  /* Get initial position. */
  run_ovm ();

  printf ("\nInitial Fuel: %f\n", get_fuel ());

  ui_data uid;
  uid.have_fs = false;
  uid.num_oth = 1;

  bool cont_sim = true;

  while (cont_sim)
  {
    run_ovm ();

    long double x = get_output (0x2);
    long double y = get_output (0x3);

    uid.x = x;
    uid.y = y;

    uid.oth_x[0] = x - (long double) get_output (0x4);
    uid.oth_y[0] = y - (long double) get_output (0x5);

    cont_sim = update_ui (&uid);
  }
}


static void
clear_skies_solver (void)
{
  set_inputs (1, SCENARIO_PORT, (double) scenario_id);

  /* Get initial position. */
  run_ovm ();

  printf ("\nInitial Fuel: %f\n", get_fuel ());

  ui_data uid;
  uid.have_fs = true;
  uid.num_oth = MAX_OTH_SATS;

  bool cont_sim = true;

  while (cont_sim)
  {
    run_ovm ();

    long double x = get_output (0x2);
    long double y = get_output (0x3);

    uid.x = x;
    uid.y = y;

    uid.fs_x = x - (long double) get_output (0x4);
    uid.fs_y = y - (long double) get_output (0x5);

    for (int i = 0; i < MAX_OTH_SATS; i++)
    {
      uid.oth_x[i] = x - (long double) get_output (3*i + 0x7);
      uid.oth_y[i] = y - (long double) get_output (3*i + 0x8);
    }

    cont_sim = update_ui (&uid);
  }
}


static void
process_args (int argc, char *argv[])
{
  /* Default problem and configuration. */
  scenario_id = 1001U;
  obf_file = "bin1.obf";
  problem_name = "Hohmann Transfer Orbit";
  problem_solver = hohmann_solver;


  if (argc > 1)
  {
    /* Custom problem and configuration. */
    char *id_str = argv[1];

    scenario_id = (uint32_t) (atoi (id_str));

    switch (scenario_id)
    {
    case 1001:
    case 1002:
    case 1003:
    case 1004:
      /* Same as the defaults. */
      break;

    case 2001:
    case 2002:
    case 2003:
    case 2004:
      obf_file = "bin2.obf";
      problem_name = "Meet and Greet";
      problem_solver = mng_solver;
      break;

    case 3001:
    case 3002:
    case 3003:
    case 3004:
      obf_file = "bin3.obf";
      problem_name = "Eccentric Meet and Greet";
      problem_solver = ecc_mng_solver;
      break;

    case 4001:
    case 4002:
    case 4003:
    case 4004:
      obf_file = "bin4.obf";
      problem_name = "Operation Clear Skies";
      problem_solver = clear_skies_solver;
      break;

    default:
      fprintf
        (stderr, "\nERROR: Invalid Scenario Identifier \"%s\"\n", id_str);
      exit (1);
      break;
    }

    if (argc > 2)
    {
      char *sz_str = argv[2];

      if (strcmp (sz_str, "big") == 0)
        es = BIG_EARTH;
      else if (strcmp (sz_str, "medium") == 0)
        es = MEDIUM_EARTH;
      else if (strcmp (sz_str, "small") == 0)
        es = SMALL_EARTH;
      else
      {
        fprintf
          (stderr, "\nERROR: Invalid Earth Size \"%s\"\n", sz_str);
        exit (1);
      }
    }
  }
}


int
main (int argc, char *argv[])
{
  process_args (argc, argv);

  load_obf (obf_file);

  init_ui (problem_name, scenario_id, es);
  init_scenario (scenario_id);

  (*problem_solver) ();

  quit_scenario ();
  quit_ui ();

  return 0;
}
