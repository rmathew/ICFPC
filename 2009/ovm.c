#include <stdint.h>
#include <stdbool.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <stdarg.h>
#include <math.h>

#include "ovm.h"

#define NUM_IO_PORTS 0x4000

#define NUM_ADDRS 0x4000

#define OSF_SIG 0xCAFEBABE


static int32_t time_step = -1;

static double inp_ports[NUM_IO_PORTS];
static double out_ports[NUM_IO_PORTS];

static uint32_t inst[NUM_ADDRS];
static double data[NUM_ADDRS];

static bool status = false;

static double fuel = 0.0;
static double score = 0.0;

static FILE *osf_fp = NULL;

static int num_insts = 0;

static union
{
  struct
  {
    double d;
    uint32_t i;
  } even;

  struct
  {
    uint32_t i;
    double d;
  } odd;
} frame;


uint32_t
get_time_step (void)
{
  return time_step;
}


double
get_fuel (void)
{
  return fuel;
}


double
get_score (void)
{
  return score;
}


double
get_output (uint32_t p_num)
{
  return out_ports[p_num];
}


void
set_inputs (uint32_t num_inp, ...)
{
  va_list args_list;

  fwrite (&time_step, sizeof (time_step), 1, osf_fp);
  fwrite (&num_inp, sizeof (num_inp), 1, osf_fp);

  va_start (args_list, num_inp);
  for (int i = 0; i < num_inp; i++)
  {
    uint32_t port_num = va_arg (args_list, uint32_t);
    double port_val = va_arg (args_list, double);

    inp_ports[port_num] = port_val;

    fwrite (&port_num, sizeof (port_num), 1, osf_fp);
    fwrite (&port_val, sizeof (port_val), 1, osf_fp);
  }
  va_end (args_list);
}


void
load_obf (const char *inp_file)
{
  FILE *fp = fopen (inp_file, "rb");

  if (fp == NULL)
  {
    perror ("\nERROR: Invalid Input File");
    exit (1);
  }

  printf ("\nLoading Orbit Binary File \"%s\"...", inp_file);

  memset (inst, 0, sizeof (inst));
  memset (data, 0, sizeof (data));

  memset (inp_ports, 0, sizeof (inp_ports));
  memset (out_ports, 0, sizeof (out_ports));

  int curr_frame = 0;
  bool even = true;

  size_t n;
  while ((n = fread (&frame, 1, sizeof (frame), fp)) > 0)
  {
    bool partial = (n < sizeof (frame));
    if (partial)
    {
      fprintf (stderr, "\nERROR: Incomplete Frame #%u\n", curr_frame);
      exit (1);
    }

    if (even)
    {
      inst[curr_frame] = frame.even.i;
      data[curr_frame] = frame.even.d;
    }
    else
    {
      inst[curr_frame] = frame.odd.i;
      data[curr_frame] = frame.odd.d;
    }

    even = !even;
    curr_frame++;
  }

  fclose (fp);

  printf (" (%u frames)\n", curr_frame);

  num_insts = curr_frame;
}


void
init_scenario (uint32_t s_id)
{
  status = false;
  time_step = -1;
  score = 0.0;

  char buf[64];
  snprintf (buf, sizeof (buf), "%u_%u.osf", TEAM_ID, s_id);

  if ((osf_fp = fopen (buf, "wb")) == NULL)
  {
    perror ("\nERROR: Invalid Output File");
    exit (1);
  }

  printf ("\nCreating Orbit Solution File \"%s\"...\n", buf);

  struct
  {
    uint32_t magic;
    uint32_t team_id;
    uint32_t scen_id;
  } hdr;

  hdr.magic = OSF_SIG;
  hdr.team_id = TEAM_ID;
  hdr.scen_id = s_id;

  fwrite (&hdr, sizeof (hdr), 1, osf_fp);
}


void
quit_scenario (void)
{
  if (osf_fp != NULL)
  {
    fclose (osf_fp);
    osf_fp = NULL;
  }
}


void
run_ovm (void)
{
  time_step++;

  for (int n = 0; n < num_insts; n++)
  {
    uint32_t an_inst = inst[n];

    uint32_t op = (an_inst & 0xF0000000U) >> 28;

    uint32_t r1 = (an_inst & 0x0FFFC000U) >> 14;
    uint32_t r2 = (an_inst & 0x00003FFFU);

    uint32_t sop = (an_inst & 0x0F000000U) >> 24;
    uint32_t imm = (an_inst & 0x00E00000U) >> 21;

    switch (op)
    {
    case 0x0:
      /* S-Type OpCode */
      /* NOTE: "r2" is "r1" for S-Type instructions. */
      switch (sop)
      {
      case 0x0:
        /* NOOP */
        break;

      case 0x1:
        /* CMPZ */
        switch (imm)
        {
        case 0x0: /* LTZ */
          status = (data[r2] < 0.0);
          break;

        case 0x1: /* LEZ */
          status = (data[r2] <= 0.0);
          break;

        case 0x2: /* EQZ */
          status = (data[r2] == 0.0);
          break;

        case 0x3: /* GEZ */
          status = (data[r2] >= 0.0);
          break;

        case 0x4: /* GTZ */
          status = (data[r2] > 0.0);
          break;

        default:
          fprintf (stderr, "\nERROR: Invalid CMPZ Immediate\n");
          exit (1);
          break;
        }
        break;

      case 0x2:
        /* SQRT */
        if (data[r2] < 0.0)
        {
          fprintf (stderr, "\nERROR: Square-Root of Negative Number\n");
          exit (1);
        }
        else
          data[n] = sqrt (data[r2]);
        break;

      case 0x3:
        /* COPY */
        data[n] = data[r2];
        break;

      case 0x4:
        /* INPUT */
        data[n] = inp_ports[r2];
        break;

      default:
        fprintf (stderr, "\nERROR: Invalid S-Type OpCode %u\n", sop);
        exit (1);
        break;
      }
      break;

    case 0x1:
      /* ADD */
      data[n] = data[r1] + data[r2];
      break;

    case 0x2:
      /* SUB */
      data[n] = data[r1] - data[r2];
      break;

    case 0x3:
      /* MULT */
      data[n] = data[r1] * data[r2];
      break;

    case 0x4:
      /* DIV */
      if (data[r2] == 0.0)
        data[n] = 0.0;
      else
        data[n] = data[r1] / data[r2];
      break;

    case 0x5:
      /* OUTPUT */
      out_ports[r1] = data[r2];
      break;

    case 0x6:
      /* PHI */
      data[n] = status ? data[r1] : data[r2];
      break;

    default:
      fprintf (stderr, "\nERROR: Invalid D-Type OpCode %u\n", op);
      exit (1);
      break;
    }
  }

  fuel = out_ports[0x1];

  double tmp_score = out_ports[0x0];
  if (score == 0.0 && tmp_score != 0.0)
  {
    score = tmp_score;

    printf
      ("Score: %f at %u seconds (Fuel Left: %f)\n", score, time_step, fuel);

    uint32_t ts = time_step + 1U;
    uint32_t ct = 0U;

    fwrite (&ts, sizeof (ts), 1, osf_fp);
    fwrite (&ct, sizeof (ct), 1, osf_fp);
  }
}
