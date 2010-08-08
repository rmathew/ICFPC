#include <stdio.h>
#include <stdlib.h>
#include <string.h>

/* The Boehm-GC library. */
#include <gc.h>

/* The cord implementation from the Boehm-GC library. */
#include <cord.h>

#include "seq.h"


#define RNA_BLOCK_SIZE 7

#define ENDO_DNA_FILE "endo.dna"

#define ENDO_RNA_FILE "endo.rna"


/* The type of a pattern item. */
typedef enum
{
  PITEM_INVALID = 0,
  PITEM_BASE,
  PITEM_SKIP,
  PITEM_SEARCH,
  PITEM_OPEN,
  PITEM_CLOSE
} pitem_type;

/* A pattern item. */
typedef struct pitem
{
  pitem_type type;
  union
  {
    char base;
    unsigned int skip_count;
    CORD srch;
  } u;
} *pitem;

/* The type of a template item. */
typedef enum
{
  TITEM_INVALID = 0,
  TITEM_BASE,
  TITEM_REF_NUM,
  TITEM_REF_LEN
} titem_type;

/* A reference number and its protection level (for template items). */
typedef struct titem_ref_num
{
  unsigned int num;
  unsigned int level;
} titem_ref_num;

/* A template item. */
typedef struct titem
{
  titem_type type;
  union
  {
    char *base;
    struct titem_ref_num ref_num;
    unsigned int ref_len;
  } u;
} *titem;


static CORD dna;

/*
  Tracks the current character of the DNA. This should not be allowed to
  become invalid and should be updated whenever the DNA is updated.
*/
static CORD_pos dna_pos;

/*
  Tracks the number of remaining DNA bases. This should be updated whenever
  a DNA base is consumed.
*/
static unsigned int dna_len;


/* Macros to keep dna_len and dna_pos updated. */

#define RESET_DNA_CURSOR() \
  CORD_set_pos (dna_pos, dna, 0); dna_len = CORD_len (dna)

#define CONSUME_DNA_BASE() CORD_next (dna_pos); dna_len--


unsigned int num_rna;
char *rna_file = ENDO_RNA_FILE;
FILE *rna_fp;
char *rna_buf;


/* Reusable DNA base CORDs. */
static CORD cord_i = "I";
static CORD cord_c = "C";
static CORD cord_f = "F";
static CORD cord_p = "P";

/* Reusable pattern items. Initialised in init(). */
static pitem pitem_base_i;
static pitem pitem_base_c;
static pitem pitem_base_f;
static pitem pitem_base_p;
static pitem pitem_open;
static pitem pitem_close;


/* Reusable template items. Initialised in init(). */
static titem titem_base_i;
static titem titem_base_c;
static titem titem_base_f;
static titem titem_base_p;

/* Whether to print out lots of debugging information or not. */
static int debug = 0;

static char *prefix_file = NULL;


/*
  Initialises some global variables.
*/
static void
init (void)
{
  GC_INIT ();

  /* Read from one or more files later. */
  dna = CORD_EMPTY;

  pitem_base_i = (pitem) GC_MALLOC_ATOMIC (sizeof (struct pitem));
  pitem_base_i->type = PITEM_BASE;
  pitem_base_i->u.base = 'I';

  pitem_base_c = (pitem) GC_MALLOC_ATOMIC (sizeof (struct pitem));
  pitem_base_c->type = PITEM_BASE;
  pitem_base_c->u.base = 'C';

  pitem_base_f = (pitem) GC_MALLOC_ATOMIC (sizeof (struct pitem));
  pitem_base_f->type = PITEM_BASE;
  pitem_base_f->u.base = 'F';

  pitem_base_p = (pitem) GC_MALLOC_ATOMIC (sizeof (struct pitem));
  pitem_base_p->type = PITEM_BASE;
  pitem_base_p->u.base = 'P';

  pitem_open = (pitem) GC_MALLOC_ATOMIC (sizeof (struct pitem));
  pitem_open->type = PITEM_OPEN;

  pitem_close = (pitem) GC_MALLOC_ATOMIC (sizeof (struct pitem));
  pitem_close->type = PITEM_CLOSE;


  titem_base_i = (titem) GC_MALLOC_ATOMIC (sizeof (struct titem));
  titem_base_i->type = TITEM_BASE;
  titem_base_i->u.base = "I";

  titem_base_c = (titem) GC_MALLOC_ATOMIC (sizeof (struct titem));
  titem_base_c->type = TITEM_BASE;
  titem_base_c->u.base = "C";

  titem_base_f = (titem) GC_MALLOC_ATOMIC (sizeof (struct titem));
  titem_base_f->type = TITEM_BASE;
  titem_base_f->u.base = "F";

  titem_base_p = (titem) GC_MALLOC_ATOMIC (sizeof (struct titem));
  titem_base_p->type = TITEM_BASE;
  titem_base_p->u.base = "P";
}

/*
  Finishes the "execute" programme by writing the RNA sequence to a file.
*/
static void
finish (void)
{
  fclose (rna_fp);
  exit (0);
}

/*
  Decodes a DNA-encoded number.
*/
static unsigned int
nat (void)
{
  unsigned int ret_val = 0;
  
  if( dna_len > 0)
  {
    char x = CORD_pos_fetch (dna_pos);
    CONSUME_DNA_BASE ();

    switch (x)
    {
    case 'P':
      ret_val = 0;
      break;

    case 'I':
    case 'F':
      ret_val = 2 * nat ();
      break;

    case 'C':
      ret_val = 2 * nat () + 1;
      break;
    
    default:
      fprintf (
        stderr, "ERROR: Unknown base '%c' in DNA at line %d!\n", x, __LINE__);
      exit (1);
      break;
    }
  }
  else
  {
    finish ();
  }

  return ret_val;
}

/*
  Decodes a sequence of bases (from the task specification).
*/
static CORD
consts (void)
{
  CORD ret_val = CORD_EMPTY;
  CORD s;

  if( dna_len > 0)
  {
    char x = CORD_pos_fetch (dna_pos);

    switch (x)
    {
    case 'C':
      CONSUME_DNA_BASE ();
      s = consts ();
      ret_val = CORD_cat (cord_i, s);
      break;

    case 'F':
      CONSUME_DNA_BASE ();
      s = consts ();
      ret_val = CORD_cat (cord_c, s);
      break;

    case 'P':
      CONSUME_DNA_BASE ();
      s = consts ();
      ret_val = CORD_cat (cord_f, s);
      break;

    case 'I':
      if( dna_len > 1)
      {
        unsigned int curr_idx = CORD_pos_to_index (dna_pos);

        x = CORD_fetch (dna, curr_idx + 1);

        if (x == 'C')
        {
          CONSUME_DNA_BASE ();
          CONSUME_DNA_BASE ();

          s = consts ();
          ret_val = CORD_cat (cord_p, s);
        }
      }
      break;
    
    default:
      fprintf (
        stderr, "ERROR: Unknown base '%c' in DNA at line %d!\n", x, __LINE__);
      exit (1);
      break;
    }
  }

  return ret_val;
}

/*
  Decodes a DNA pattern.
*/
static seq
pattern (void)
{
  seq ret_seq = seq_new ();

  int lvl = 0;
  int done = 0;

  while (done == 0)
  {
    if (dna_len > 0)
    {
      char x = CORD_pos_fetch (dna_pos);
      
      switch (x)
      {
      case 'C':
        CONSUME_DNA_BASE ();
        seq_append (ret_seq, pitem_base_i);
        break;

      case 'F':
        CONSUME_DNA_BASE ();
        seq_append (ret_seq, pitem_base_c);
        break;

      case 'P':
        CONSUME_DNA_BASE ();
        seq_append (ret_seq, pitem_base_f);
        break;

      case 'I':
        if (dna_len > 1)
        {
          CONSUME_DNA_BASE ();

          x = CORD_pos_fetch (dna_pos);
          
          switch (x)
          {
          case 'C':
            CONSUME_DNA_BASE ();
            seq_append (ret_seq, pitem_base_p);
            break;

          case 'P':
            {
              unsigned n;
              pitem skip;

              CONSUME_DNA_BASE ();
              n = nat ();

              skip = (pitem) GC_MALLOC_ATOMIC (sizeof (struct pitem));
              skip->type = PITEM_SKIP;
              skip->u.skip_count = n;
              seq_append (ret_seq, skip);
            }
            break;

          case 'F':
            {
              pitem search;
              CORD srch_dna;

              /* Note: An extra base is consumed. */
              CONSUME_DNA_BASE ();
              CONSUME_DNA_BASE ();

              srch_dna = consts ();

              search = (pitem) GC_MALLOC (sizeof (struct pitem));
              search->type = PITEM_SEARCH;
              search->u.srch = srch_dna;
              seq_append (ret_seq, search);
            }
            break;

          case 'I':
            if (dna_len > 2)
            {
              CONSUME_DNA_BASE ();

              x = CORD_pos_fetch (dna_pos);
              
              switch (x)
              {
              case 'P':
                CONSUME_DNA_BASE ();
                lvl += 1;
                seq_append (ret_seq, pitem_open);
                break;

              case 'C':
              case 'F':
                CONSUME_DNA_BASE ();
                if (lvl == 0)
                  return ret_seq;
                else
                {
                  lvl -= 1;
                  seq_append (ret_seq, pitem_close);
                }
                break;

              case 'I':
                {
                  unsigned int i;

                  CONSUME_DNA_BASE ();
                  for (i = 0; i < RNA_BLOCK_SIZE; i++)
                  {
                    x = CORD_pos_fetch (dna_pos);
                    fputc (x, rna_fp);
                    CONSUME_DNA_BASE ();
                  }
                  fputc ('\n', rna_fp);
                  num_rna++;
                }
                break;
              
              default:
                fprintf (
                  stderr, "ERROR: Unknown base '%c' in DNA at line %d!\n", x,
                  __LINE__);
                exit (1);
                break;
              }
            }
            else
            {
              done = 1;
              finish ();
            }
            break;
          
          default:
            fprintf (
              stderr, "ERROR: Unknown base '%c' in DNA at line %d!\n", x,
              __LINE__);
            exit (1);
            break;
          }
        }
        else
        {
          done = 1;
          finish ();
        }
        break;
      
      default:
        fprintf (
          stderr, "ERROR: Unknown base '%c' in DNA at line %d!\n", x,
          __LINE__);
        exit (1);
        break;
      }
    }
    else
    {
      done = 1;
      finish ();
    }
  }

  return ret_seq;
}

/*
  Prints a DNA pattern to the given stream. Useful for debugging.
*/
static void
print_pattern (seq s, FILE *fp)
{
  seq_iter si = seq_get_iter (s);
  while (seq_iter_has_next (si))
  {
    pitem p = (pitem) seq_iter_next (si);

    if (p->type == PITEM_BASE)
      fputc (p->u.base, fp);
    else if (p->type == PITEM_OPEN)
      fputc ('(', fp);
    else if (p->type == PITEM_CLOSE)
      fputc (')', fp);
    else if (p->type == PITEM_SKIP)
      fprintf (fp, "!%u", p->u.skip_count);
    else if (p->type == PITEM_SEARCH)
      fprintf (fp, "?%s", CORD_to_char_star (p->u.srch));
  }
  seq_close_iter (si);
}

/*
  Decodes a DNA template.
*/
static seq
template (void)
{
  seq ret_seq = seq_new ();

  int done = 0;

  while (done == 0)
  {
    if (dna_len > 0)
    {
      char x = CORD_pos_fetch (dna_pos);

      switch (x)
      {
      case 'C':
        CONSUME_DNA_BASE ();
        seq_append (ret_seq, titem_base_i);
        break;

      case 'F':
        CONSUME_DNA_BASE ();
        seq_append (ret_seq, titem_base_c);
        break;

      case 'P':
        CONSUME_DNA_BASE ();
        seq_append (ret_seq, titem_base_f);
        break;

      case 'I':
        if (dna_len > 1)
        {
          CONSUME_DNA_BASE ();

          x = CORD_pos_fetch (dna_pos);

          switch (x)
          {
          case 'C':
            CONSUME_DNA_BASE ();
            seq_append (ret_seq, titem_base_p);
            break;
            
          case 'F':
          case 'P':
            {
              unsigned int l, n;
              titem ref_num;

              CONSUME_DNA_BASE ();

              l = nat ();
              n = nat ();

              ref_num = (titem) GC_MALLOC_ATOMIC (sizeof (struct titem));
              ref_num->type = TITEM_REF_NUM;
              ref_num->u.ref_num.num = n;
              ref_num->u.ref_num.level = l;

              seq_append (ret_seq, ref_num);
            }
            break;
            
          case 'I':
            if (dna_len > 2)
            {
              CONSUME_DNA_BASE ();

              x = CORD_pos_fetch (dna_pos);

              switch (x)
              {
              case 'C':
              case 'F':
                CONSUME_DNA_BASE ();
                return ret_seq;
                break;

              case 'P':
                {
                  unsigned int n;
                  titem ref_len;

                  CONSUME_DNA_BASE ();

                  n = nat ();

                  ref_len = (titem) GC_MALLOC_ATOMIC (sizeof (struct titem));
                  ref_len->type = TITEM_REF_LEN;
                  ref_len->u.ref_len = n;
                  seq_append (ret_seq, ref_len);
                }
                break;

              case 'I':
                {
                  unsigned int i;

                  CONSUME_DNA_BASE ();
                  for (i = 0; i < RNA_BLOCK_SIZE; i++)
                  {
                    x = CORD_pos_fetch (dna_pos);
                    fputc (x, rna_fp);
                    CONSUME_DNA_BASE ();
                  }
                  fputc ('\n', rna_fp);
                }
                break;
              
              default:
                fprintf (
                  stderr, "ERROR: Unknown base '%c' in DNA at line %d!\n", x,
                  __LINE__);
                exit (1);
                break;
              }
            }
            else
            {
              done = 1;
              finish ();
            }
            break;
          
          default:
            fprintf (
              stderr, "ERROR: Unknown base '%c' in DNA at line %d!\n", x,
              __LINE__);
            exit (1);
            break;
          }
        }
        else
        {
          done = 1;
          finish ();
        }
        break;
      
      default:
        fprintf (
          stderr, "ERROR: Unknown base '%c' in DNA at line %d!\n", x,
          __LINE__);
        exit (1);
        break;
      }
    }
    else
    {
      done = 1;
      finish ();
    }
  }

  return ret_seq;
}

/*
  Prints a DNA template to the given stream. Useful for debugging.
*/
static void
print_template (seq s, FILE *fp)
{
  seq_iter si = seq_get_iter (s);
  while (seq_iter_has_next (si))
  {
    titem t = (titem) seq_iter_next (si);

    if (t->type == TITEM_BASE)
      fputs (t->u.base, fp);
    else if (t->type == TITEM_REF_NUM)
    {
      unsigned int i = t->u.ref_num.level;

      fputc ('\\', fp);
      while (i > 0)
      {
        fputc ('\\', fp);
        i--;
      }

      fprintf (fp, "%u", t->u.ref_num.num);
    }
    else if (t->type == TITEM_REF_LEN)
      fprintf (fp, "|%u|", t->u.ref_len);
  }
  seq_close_iter (si);
}

/*
  Encodes a number as DNA bases.
*/
static CORD
asnat (unsigned int n)
{
  /* Note: N can have at most 32 bits. */
  char ret_val[34];

  unsigned int i = 0;

  while (n > 0)
  {
    if (n & 1)
      ret_val[i++] = 'C';
    else
      ret_val[i++] = 'I';

    n >>= 1;
  }
  ret_val[i++] = 'P';
  ret_val[i] = '\0';

  return CORD_from_char_star (ret_val);
}

/*
  Quotes a DNA sequence (from the specification).
*/
static CORD
quote (CORD s)
{
  /*
    Note: The worst case is a string of 'P' bases each of which converts to
    an 'IC' sequence.
  */
  char *ret_val = GC_MALLOC_ATOMIC (2 * CORD_len (s) + 1);

  char *tmp;

  tmp = ret_val;
  CORD_pos pos;
  CORD_FOR (pos, s)
  {
    char x = CORD_pos_fetch (pos);

    switch (x)
    {
    case 'I':
      *tmp++ = 'C';
      break;

    case 'C':
      *tmp++ = 'F';
      break;

    case 'F':
      *tmp++ = 'P';
      break;

    case 'P':
      *tmp++ = 'I';
      *tmp++ = 'C';
      break;
    
    default:
      fprintf (
        stderr, "ERROR: Unknown base '%c' in DNA at line %d!\n", x,
        __LINE__);
      exit (1);
      break;
    }
  }
  *tmp = '\0';

  return CORD_from_char_star (ret_val);
}

static CORD
protect (unsigned int lvl, CORD d)
{
  if (lvl == 0)
    return d;
  else
  {
    return protect (lvl - 1, quote (d));
  }
}

/*
  Replaces DNA according to a given template (from the specification).
*/
static void
replace (seq tpl, seq env)
{
  CORD x;
  CORD y;

  CORD r = CORD_EMPTY;

  seq_iter si = seq_get_iter (tpl);
  while (seq_iter_has_next (si))
  {
    titem ti = (titem) seq_iter_next (si);

    switch (ti->type)
    {
    case TITEM_BASE:
      r = CORD_cat_char_star (r, ti->u.base, 1);
      break;

    case TITEM_REF_NUM:
      x = seq_elem_at (env, ti->u.ref_num.num);
      y = protect (ti->u.ref_num.level, x);
      r = CORD_cat (r, y);
      break;

    case TITEM_REF_LEN:
      x = seq_elem_at (env, ti->u.ref_num.num);
      y = asnat (CORD_len (x));
      r = CORD_cat (r, y);
      break;

    default:
      break;
    }
  }
  seq_close_iter (si);

  dna = CORD_cat (r, CORD_substr (dna, CORD_pos_to_index (dna_pos), dna_len));
  RESET_DNA_CURSOR ();
}

/*
  Matches and replaces DNA (from the specification).
*/
static void
matchreplace (seq pat, seq tpl)
{
  unsigned int i = 0;

  seq e = seq_new ();
  seq c = seq_new ();

  unsigned int curr_idx = CORD_pos_to_index (dna_pos);
  unsigned int *tmp_ptr;
  unsigned int tmp;
  char x;

  CORD tmp_dna;

  seq_iter si = seq_get_iter (pat);
  while (seq_iter_has_next (si))
  {
    pitem pi = (pitem) seq_iter_next (si);

    switch (pi->type)
    {
    case PITEM_BASE:
      x = CORD_fetch (dna, curr_idx + i);
      if (x == pi->u.base)
        i++;
      else
      {
        if (debug)
          printf ("Failed match.\n");

        return;
      }
      break;

    case PITEM_SKIP:
      i += pi->u.skip_count;
      if (i > dna_len)
      {
        if (debug)
          printf ("Failed match.\n");

        return;
      }
      break;

    case PITEM_SEARCH:
      if ((tmp = CORD_str (dna, curr_idx + i, pi->u.srch)) != CORD_NOT_FOUND)
        i = CORD_len (pi->u.srch) + (tmp - curr_idx);
      else
      {
        if (debug)
          printf ("Failed match.\n");

        return;
      }
      break;

    case PITEM_OPEN:
      tmp_ptr
        = (unsigned int *) GC_MALLOC_ATOMIC (sizeof (unsigned int));
      *tmp_ptr = i;
      seq_prepend (c, tmp_ptr);
      break;

    case PITEM_CLOSE:
      tmp_ptr = (unsigned int *) seq_elem_at (c, 0);
      tmp = *tmp_ptr;
      tmp_dna = CORD_substr (dna, curr_idx + tmp, i - tmp);
      seq_append (e, (void *) tmp_dna);
      seq_discard (c, 1);
      break;

    default:
      break;
    }
  }
  seq_close_iter (si);


  if (debug)
  {
    if (i > 0)
    {
      unsigned int n = 0;

      printf ("Successful match of length %u.\n", i);
      seq_iter ei = seq_get_iter (e);
      while (seq_iter_has_next (ei))
      {
        CORD ds = seq_iter_next (ei);

        printf (
          "e[%u] has %u bases (%s).\n", n, CORD_len (ds),
          (CORD_len (ds) <= 30) ? CORD_to_char_star (ds) : "...");

        n++;
      }
      seq_close_iter (ei);
    }
    else
      printf ("Failed match.\n");
  }
    
  dna = CORD_substr (dna, curr_idx + i, dna_len - i);
  RESET_DNA_CURSOR ();

  replace (tpl, e);
}

/*
  Reads a strand of DNA from a given file.
*/
static void
read_dna (char *f)
{
  CORD r;

  printf ("\nReading DNA from file \"%s\"...", f);
  fflush (stdout);

  r = CORD_from_file (fopen (f, "r"));
  dna = CORD_cat (dna, r);
  RESET_DNA_CURSOR ();

  printf ("done.\n");
}

static void
open_rna_file (void)
{
  rna_fp = fopen (rna_file, "w");
  if (rna_fp == NULL)
  {
    fprintf
      (stderr, "ERROR: Could not open RNA file \"%s\" for writing!\n",
      ENDO_RNA_FILE);

    exit (1);
  }
  else
  {
    rna_buf = GC_MALLOC_ATOMIC (sizeof (char) * (RNA_BLOCK_SIZE + 1) * 100000);
    setbuf (rna_fp, rna_buf);

    printf ("\nOutput RNA will be written to file \"%s\"...\n", rna_file);
  }
  num_rna = 0;
}

/*
  Executes a DNA strand to create RNA.
*/
static void
execute (void)
{
  unsigned int cycle = 0;

  dna = CORD_EMPTY;

  if (prefix_file != NULL)
    read_dna (prefix_file);

  read_dna (ENDO_DNA_FILE);

  open_rna_file ();

  if (!debug)
  {
    printf ("\nMunching: \n");
    fflush (stdout);
  }

  while (1)
  {
    if (debug)
    {
      printf ("\nIteration %u with %u DNA bases...\n", cycle, dna_len);
      fflush (stdout);
    }
    else
    {
      if ((cycle % 100000) == 0)
      {
        printf ("Iteration #%u (%u RNA commands)\n", cycle, num_rna);
        fflush (stdout);
      }
    }


    seq pat = pattern ();

    if (debug)
    {
      printf ("Pattern: ");
      print_pattern (pat, stdout);
      printf ("\n");
    }


    seq tpl = template ();

    if (debug)
    {
      printf ("Template: ");
      print_template (tpl, stdout);
      printf ("\n");
    }


    matchreplace (pat, tpl);

    if (debug)
    {
      printf ("RNA now has %u commands.\n", num_rna);

      if (cycle == 99)
        exit (0);
    }

    cycle++;
  }
}

/*
  Some tests from the specification to see if our implementation is sane.
*/
static void
test_sanity (void)
{
  char *init_dnas[] = {
    "IIPIPICPIICICIIFICCIFPPIICCFPC",
    "IIPIPICPIICICIIFICCIFCCCPPIICCFPC",
    "IIPIPIICPIICIICCIICFCFC",
  };

  char *res_dnas[] = {
    "PICFC",
    "PIICCFCFFPC",
    "I",
  };

  unsigned int num_tests = (sizeof (init_dnas) / sizeof (init_dnas[0]));
  unsigned int i, curr_idx;

  printf ("Testing sanity...\n");
  for (i = 0; i < num_tests; i++)
  {
    seq pat, tpl;
    char *dna_str;

    dna = CORD_from_char_star (init_dnas[i]);
    RESET_DNA_CURSOR ();

    pat = pattern ();

    tpl = template ();

    if (debug)
    {
      printf ("\nTest #%u: Initial DNA \"%s\"\n", i, init_dnas[i]);

      printf ("Pattern: ");
      print_pattern (pat, stdout);
      printf ("\n");

      printf ("Template: ");
      print_template (tpl, stdout);
      printf ("\n");
    }

    matchreplace (pat, tpl);

    curr_idx = CORD_pos_to_index (dna_pos);
    dna_str = CORD_to_char_star (CORD_substr (dna, curr_idx, dna_len));

    if (debug)
    {
      printf (
        "Test #%u: Resultant DNA \"%s\", Expected DNA \"%s\"\n", i, dna_str,
        res_dnas[i]);
    }

    printf ("Test #%u: ", i);
    if (strcmp (dna_str, res_dnas[i]) == 0)
      printf ("PASSED.\n");
    else
    {
      printf ("FAILED!\n");
      exit (1);
    }
  }
}

/*
  The entry-point of the programme.
*/
int
main (int argc, char *args[])
{
  if( argc > 1)
  {
    int i;
    for (i = 1; i < argc; i++)
    {
      if (strcmp (args[i], "-d") == 0)
        debug = 1;
      else if (strcmp (args[i], "-p") == 0)
      {
        i++;
        if (i == argc)
        {
          fprintf
            (stderr,
            "ERROR: \"-p\" needs the file-name of the prefix DNA!\n");

          exit (1);
        }
        else
          prefix_file = args[i];
      }
      else if (strcmp (args[i], "-o") == 0)
      {
        i++;
        if (i == argc)
        {
          fprintf
            (stderr,
            "ERROR: \"-o\" needs the file-name of the output RNA!\n");

          exit (1);
        }
        else
          rna_file = args[i];
      }
    }
  }

  init ();

  test_sanity ();

  execute ();

  return 0;
}
