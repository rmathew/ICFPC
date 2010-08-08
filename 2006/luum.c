/*
 * ICFPC 2006
 * ==========
 *
 * Universal Machine (UM) emulator by Ranjit Mathew <rmathew@gmail.com>
 * (team "Codermal"). (Less ugly version.)
 *
 * Usage: um [<scroll file>]
 *
 * The default for scroll file is "codex.umz".
 */
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <errno.h>
#include <sys/types.h>
#include <sys/stat.h>
#include <unistd.h>


/* The default programme scroll used by the UM. */
#define DEFAULT_SCROLL "codex.umz"

/* The number of bytes per platter. */
#define PLATTER_BYTES 4

/* The number of operators supported by the UM. */
#define NUM_OPS 14

/* The number of registers in the UM. */
#define NUM_REGS 8

/* The block-size in which the "array of arrays" is incremented. */
#define ARRAYS_BLOCK 10000

/* Convert a big-endian platter encoding from a scroll into a platter. */
#define MAKE_PLATTER(a,b,c,d) \
  ((((platter_t)(a) & 0x000000FF) << 24) \
   | (((platter_t)(b) & 0x000000FF) << 16) \
   | (((platter_t)(c) & 0x000000FF) << 8) \
   | (((platter_t)(d) & 0x000000FF)))

/* Get the "operator" field of an instruction platter. */
#define GET_OP(x) ((unsigned int)((x) & 0xF0000000) >> 28)

/* Get the "register A" field of an instruction platter. */
#define GET_REG_A(x) ((unsigned int)((x) & 0x000001C0) >> 6)

/* Get the "register B" field of an instruction platter. */
#define GET_REG_B(x) ((unsigned int)((x) & 0x00000038) >> 3)

/* Get the "register C" field of an instruction platter. */
#define GET_REG_C(x) ((unsigned int)((x) & 0x00000007))


/* A platter. Should be a 32-bit unsigned integer. */
typedef unsigned int platter_t;

/* An array. */
typedef struct array
{
  unsigned int platters;
  platter_t *array;
} array_t;

/* A note for remembering the index of a freed array. */
typedef struct freed_array
{
  unsigned int idx;
  struct freed_array *next;
} freed_array_t;


/* The registers of the UM. */
static platter_t regs[NUM_REGS];

/* The arrays in the UM. */
static array_t **arrays = NULL;

/* The maximum number of arrays that have been used so far. */
static unsigned int num_arrays = 0;

/* The maximum number of arrays that can be used without forcing a realloc. */
static unsigned int max_arrays = 0;

/* A list of freed-up array indices. */
static freed_array_t *freed_arrays = NULL;

/* A pointer to the current instruction platter in the 0-th array. */
static platter_t *insn_ptr;

/* The current instruction platter. */
static platter_t insn;

/* The current operation code and operand register codes, if any. */
static unsigned int op, ra = 0, rb = 0, rc = 0;

/* The values contained in the operand registers, if any. */
static platter_t reg_a = 0, reg_b = 0, reg_c = 0;

/* The path to the file containing the current programme scroll. */
static char *scroll;

/* The signature for functions that execute operators. */
typedef int (*exec_fn_t)(void);


/* A simple wrapper around malloc() that exits with an error message if it
   could not allocate the desired amount of memory.
   Returns a pointer to N bytes of newly-allocated memory. */
static void *
emalloc( size_t n)
{
  void *ret_val = malloc( n);
  if( ret_val == NULL)
  {
    fprintf( stderr, "ERROR: Could not allocate %u bytes of memory!\n", n);
    exit( 1);
  }

  return ret_val;
}


/* Allocates memory for a new array for PLATTERS number of platters. Clears
   the array if CLEAR is set to 1. Returns a pointer to the newly-allocated
   array. Inserts a sentinel instruction platter with an illegal operator
   so that we do not have to constantly check the range of the execution
   finger. */
static array_t *
allocate( unsigned int platters, int clear)
{
  array_t *ret_val;
  platter_t *array_mem;
  unsigned int array_mem_size;

  ret_val = (array_t *)emalloc( sizeof( array_t));

  array_mem_size = (platters + 1) * sizeof( platter_t);
  array_mem = (platter_t *)emalloc( array_mem_size);
  if( clear)
  {
    memset( array_mem, 0, array_mem_size - sizeof( platter_t));
  }

  /* Insert a sentinel invalid instruction at the end of the array. */
  array_mem[platters] = 0xFFFFFFFF;

  ret_val->platters = platters;
  ret_val->array = array_mem;

  return ret_val;
}


/* Abandons an array ARRAY allocated with allocate(). */
static void
abandon( array_t *array)
{
  free( array->array);
  free( array);
}


/* Initialises the UM. Reads in the programme scroll. */
static void
initialise( void)
{
  FILE *fp;
  int i, ch1;
  struct stat st;

  /* Quick check to see if we are indeed on a system with 32-bit platters. */
  if( sizeof( platter_t) != PLATTER_BYTES)
  {
    fprintf( stderr, "ERROR: Incompatible system.\n");
    exit( 1);
  }

  /* Verify that the scroll file exists. */
  if( stat( scroll, &st) != 0)
  {
    fprintf( stderr, "ERROR: Could not open codex file \"%s\"!\n", scroll);
    perror( "ERROR");
    exit( 1);
  }

  /* Verify that the scroll file has a size divisible by the size of a
     platter. */
  if( (st.st_size % PLATTER_BYTES) != 0)
  {
    fprintf( stderr, "ERROR: Codex file possibly corrupted!\n");
    exit( 1);
  }


  /* Read in the platters from the scroll file and initialise the 0-th
     array with it. */

  num_arrays = 1;
  max_arrays = ARRAYS_BLOCK;
  arrays = (array_t **)emalloc( ARRAYS_BLOCK * sizeof( array_t *));
  arrays[0] = allocate( st.st_size / PLATTER_BYTES, 0);

  if( (fp = fopen( scroll, "r")) == NULL)
  {
    fprintf( stderr, "ERROR: Could not open codex file \"%s\"!\n", scroll);
    perror( "ERROR");
    exit( 1);
  }

  for( i = 0; (ch1 = fgetc( fp)) != EOF; i++)
  {
    int ch2, ch3, ch4;
    ch2 = fgetc( fp);
    ch3 = fgetc( fp);
    ch4 = fgetc( fp);
    arrays[0]->array[i] = MAKE_PLATTER(ch1, ch2, ch3, ch4);
  }

  fclose( fp);
}


/* Executes a Conditional Move instruction. Returns 1 if the UM should
   quit. */
static int
exec_cmov( void)
{
  rc = GET_REG_C(insn);

  if( regs[rc] != 0)
  {
    ra = GET_REG_A(insn);
    rb = GET_REG_B(insn);

    regs[ra] = regs[rb];
  }

  return 0;
}


/* Executes an Array Index instruction. Returns 1 if the UM should quit. */
static int
exec_aidx( void)
{
  int quit = 0;

  rb = GET_REG_B(insn);
  reg_b = regs[rb];

  if( reg_b >= num_arrays)
  {
    fprintf( stderr, "ERROR: OP01 - Array index out of bounds!\n");
    quit = 1;
  }
  else if( arrays[reg_b] == NULL)
  {
    fprintf( stderr, "ERROR: OP01 - Accessing abandoned array!\n");
    quit = 1;
  }
  else
  {
    rc = GET_REG_C(insn);
    reg_c = regs[rc];

    if( reg_c >= arrays[reg_b]->platters)
    {
      fprintf( stderr, "ERROR: OP01 - Platter index out of bounds!\n");
      quit = 1;
    }
    else
    {
      ra = GET_REG_A(insn);
      regs[ra] = arrays[reg_b]->array[reg_c];
    }
  }

  return quit;
}


/* Executes an Array Amendment instruction. Returns 1 if the UM should
   quit. */
static int
exec_amod( void)
{
  int quit = 0;

  ra = GET_REG_A(insn);
  reg_a = regs[ra];

  if( reg_a >= num_arrays)
  {
    fprintf( stderr, "ERROR: OP02 - Array index out of bounds!\n");
    quit = 1;
  }
  else if( arrays[reg_a] == NULL)
  {
    fprintf( stderr, "ERROR: OP02 - Accessing abandoned array!\n");
    quit = 1;
  }
  else
  {
    rb = GET_REG_B(insn);
    reg_b = regs[rb];

    if( reg_b >= arrays[reg_a]->platters)
    {
      fprintf( stderr, "ERROR: OP02 - Platter index out of bounds!\n");
      quit = 1;
    }
    else
    {
      rc = GET_REG_C(insn);

      arrays[reg_a]->array[reg_b] = regs[rc];
    }
  }

  return quit;
}


/* Executes an Addition instruction. Returns 1 if the UM should quit. */
static int
exec_addn( void)
{
  ra = GET_REG_A(insn);
  rb = GET_REG_B(insn);
  rc = GET_REG_C(insn);

  regs[ra] = regs[rb] + regs[rc];

  return 0;
}


/* Executes a Multiplication instruction. Returns 1 if the UM should quit. */
static int
exec_mult( void)
{
  ra = GET_REG_A(insn);
  rb = GET_REG_B(insn);
  rc = GET_REG_C(insn);

  regs[ra] = regs[rb] * regs[rc];

  return 0;
}


/* Executes a Division instruction. Returns 1 if the UM should quit. */
static int
exec_divn( void)
{
  int quit = 0;

  rc = GET_REG_C(insn);
  reg_c = regs[rc];

  if( reg_c == 0)
  {
    fprintf( stderr, "ERROR: OP05 - Division by zero!\n");
    quit = 1;
  }
  else
  {
    ra = GET_REG_A(insn);
    rb = GET_REG_B(insn);

    regs[ra] = regs[rb] / reg_c;
  }

  return quit;
}


/* Executes a Not-AND instruction. Returns 1 if the UM should quit. */
static int
exec_nand( void)
{
  ra = GET_REG_A(insn);
  rb = GET_REG_B(insn);
  rc = GET_REG_C(insn);

  regs[ra] = ~(regs[rb] & regs[rc]);

  return 0;
}


/* Executes a HALT instruction. Returns 1 to indicate that the UM should
   quit. */
static int
exec_halt( void)
{
  return 1;
}


/* Executes an Allocation instruction. Returns 1 if the UM should quit. */
static int
exec_allc( void)
{
  unsigned int idx;

  /* Reuse the index of the last array that was abandoned, if any. */
  if( freed_arrays != NULL)
  {
    freed_array_t *f = freed_arrays;
    freed_arrays = f->next;
    idx = f->idx;
    free( f);
  }
  else
  {
    /* No free slot available. Add new array at the end. */
    idx = num_arrays;
  }

  if( idx == num_arrays)
  {
    /* If the "array of arrays" has reached its size limit, expand it. */
    if( num_arrays == max_arrays)
    {
      arrays
        = (array_t **)realloc( arrays,
                              (max_arrays + ARRAYS_BLOCK)
                               * sizeof( array_t *));

      if( arrays == NULL)
      {
        fprintf( stderr,
                 "ERROR: Could not allocate additional arrays!\n");
        exit( 1);
      }

      max_arrays += ARRAYS_BLOCK;
    }

    num_arrays++;
  }
  
  rc = GET_REG_C(insn);
  arrays[idx] = allocate( regs[rc], 1);

  rb = GET_REG_B(insn);
  regs[rb] = (platter_t )idx;

  return 0;
}


/* Executes an Abandonment instruction. Returns 1 if the UM should quit. */
static int
exec_aban( void)
{
  int quit = 0;

  rc = GET_REG_C(insn);
  reg_c = regs[rc];

  if( reg_c == 0)
  {
    fprintf( stderr, "ERROR: OP09 - Cannot abandon 0-th array!\n");
    quit = 1;
  }
  else if( reg_c >= num_arrays)
  {
    fprintf( stderr, "ERROR: OP09 - Array index out of bounds!\n");
    quit = 1;
  }
  else if( arrays[reg_c] == NULL)
  {
    fprintf( stderr, "ERROR: OP09 - Array already abandoned!\n");
    quit = 1;
  }
  else
  {
    freed_array_t *f;

    abandon( arrays[reg_c]);
    arrays[reg_c] = NULL;

    /* Remember the slot that was holding this array. */
    f = emalloc( sizeof( freed_array_t));
    f->idx = reg_c;
    f->next = freed_arrays;
    freed_arrays = f;
  }

  return quit;
}


/* Executes an Output instruction. Returns 1 if the UM should quit. */
static int
exec_outp( void)
{
  int quit = 0;

  rc = GET_REG_C(insn);
  reg_c = regs[rc];

  if( reg_c > 255)
  {
    fprintf( stderr, "ERROR: OP10 - Output value too large!\n");
    quit = 1;
  }
  else
  {
    fputc( reg_c & 0x000000FF, stdout);
    fflush( stdout);
  }

  return quit;
}


/* Executes an Input instruction. Returns 1 if the UM should quit. */
static int
exec_inpt( void)
{
  int i;

  rc = GET_REG_C(insn);

  i = fgetc( stdin);
  if( i == EOF)
    regs[rc] = 0xFFFFFFFF;
  else
    regs[rc] = (i & 0x000000FF);

  return 0;
}


/* Executes a Load Programme instruction. Returns 1 if the UM should
   quit. */
static int
exec_load( void)
{
  int quit = 0;

  rb = GET_REG_B(insn);
  reg_b = regs[rb];

  if( reg_b >= num_arrays)
  {
    fprintf( stderr, "ERROR: OP12 - Array index out of bounds!\n");
    quit = 1;
  }
  else if( arrays[reg_b] == NULL)
  {
    fprintf( stderr, "ERROR: OP12 - Accessing abandoned array!\n");
    quit = 1;
  }
  else
  {
    if( reg_b != 0)
    {
      abandon( arrays[0]);
      arrays[0] = allocate( arrays[reg_b]->platters, 0);
      memcpy( arrays[0]->array, arrays[reg_b]->array,
              arrays[0]->platters * sizeof( platter_t));
    }

    rc = GET_REG_C(insn);
    insn_ptr = arrays[0]->array + regs[rc];
  }

  return quit;
}


/* Executes an Orthography instruction. Returns 1 if the UM should quit. */
static int
exec_orth( void)
{
  ra = ((unsigned int)(insn & 0x0E000000) >> 25);
  regs[ra] = (unsigned int)insn & 0x01FFFFFF;
  return 0;
}


/* Executes the instructions from the programme scroll, one instruction
   per spin cycle. Returns back to the caller if either the UM halts or
   there was an exception. */
static void
execute( void)
{
  unsigned int quit = 0;

  /* A table of function-pointers used to execute the instruction with
     the respective operation code. */
  exec_fn_t exec_fns[NUM_OPS]
    = {
        exec_cmov, exec_aidx, exec_amod, exec_addn, exec_mult, exec_divn,
        exec_nand, exec_halt, exec_allc, exec_aban, exec_outp, exec_inpt,
        exec_load, exec_orth,
      };

  /* Clear registers. */
  memset( regs, 0, NUM_REGS * sizeof( platter_t));

  /* Initialise execution finger. */
  insn_ptr = arrays[0]->array;

  /* Execute spin cycles till a HALT or an error. */
  while( !quit)
  {
    insn = *insn_ptr;
    insn_ptr++;

    op = GET_OP(insn);

    if( op < NUM_OPS)
    {
      quit = exec_fns[op]();
    }
    else
    {
      fprintf( stderr, "ERROR: Unknown instruction!\n");
      quit = 1;
    }
  }
}


/* Cleans up after the execution of the UM. */
static void
finish( void)
{
  /* TODO */
}


/* The entry-point into the UM programme. ARGC is the number of
   command-line arguments and ARGV is an array of strings containing the
   arguments. Returns 0 on successful execution, 1 otherwise.
   
   Only recognises a single argument, as of now, that represents the path
   to the file containing the programme scroll. */
int
main( int argc, char *argv[])
{
  scroll = DEFAULT_SCROLL;
  if( argc == 2)
  {
    scroll = argv[1];
  }

  initialise( );
  execute( );
  finish( );

  return 0;
}
