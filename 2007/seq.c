#include <stdio.h>
#include <stdlib.h>

#include <gc.h>

#include "seq.h"

/*
  Dumb sequence implementation using "cons"-style cells (Lisp) to connect
  nodes.

  A sequence should never be NULL (but can be empty, i.e. "size == 0").
  The "size" element should always equal the number of "cons" cells
  in the sequence.
*/

/*
  Creates an empty sequence.
*/
seq
seq_new (void)
{
  seq ret_val = (seq) GC_MALLOC (sizeof (struct seq));

  ret_val->size = 0;
  ret_val->head = NULL;
  ret_val->tail = NULL;

  return ret_val;
}

/*
  Deletes the given sequence.
*/
void
seq_del (seq s)
{
  struct cons *curr = s->head;

  while (curr != NULL)
  {
    struct cons *tmp = curr->cdr;
    curr->car = curr->cdr = NULL;
    curr = tmp;
  }

  s->size = 0;
  s->head = s->tail = NULL;
}

/*
  Appends the given element at the end of the given sequence.
*/
void
seq_append (seq s, void *x)
{
  struct cons *b = (struct cons *) GC_MALLOC (sizeof (struct cons));
  b->car = x;
  b->cdr = NULL;

  if (s->head == NULL)
    s->head = b;

  if (s->tail != NULL)
    s->tail->cdr = b;

  s->tail = b;
  s->size++;
}

/*
  Prepends the given element at the beginning of the given sequence.
*/
void
seq_prepend (seq s, void *x)
{
  struct cons *b = (struct cons *) GC_MALLOC (sizeof (struct cons));
  b->car = x;
  b->cdr = s->head;

  s->head = b;

  if (s->tail == NULL)
    s->tail = b;

  s->size++;
}

/*
  Returns the size of the given sequence.
*/
unsigned int
seq_size (seq s)
{
  return s->size;
}

/*
  Returns the element at the given index for the given sequence, NULL, if
  the index is out of bounds.
*/
void *
seq_elem_at (seq s, unsigned int i)
{
  void *ret_val = NULL;

  if (i < s->size)
  {
    struct cons *curr = s->head;
    unsigned int n = 0;

    while (n < i)
    {
      curr = curr->cdr;
      n++;
    }

    ret_val = curr->car;
  }

  return ret_val;
}

/*
  Discards a given number of elements from beginning of the sequence.
*/
void
seq_discard (seq s, unsigned int n)
{
  if (n <= s->size)
  {
    struct cons *curr = s->head;
    unsigned int i;
    for (i = n; i > 0; i--)
    {
      struct cons *tmp = curr->cdr;
      curr->car = curr->cdr = NULL;
      curr = tmp;
    }

    s->head = curr;
    if (s->head == NULL)
      s->tail = NULL;

    s->size -= n;
  }
}

/*
  Gets an iterator for iterating over the elements of a sequence.
*/
seq_iter
seq_get_iter (seq s)
{
  seq_iter ret_val = (seq_iter) GC_MALLOC (sizeof (struct cons *));
  *ret_val = s->head;
  return ret_val;
}

/*
  Tests if the given iterator can be used to fetch any more elements
  from the corresponding sequence.
*/
int
seq_iter_has_next (seq_iter si)
{
  return (*si != NULL);
}

/*
  Fetches the next element from the corresponding sequence in the current
  iteration.
*/
void *
seq_iter_next (seq_iter si)
{
  void *ret_val = (*si)->car;
  *si = (*si)->cdr;
  return ret_val;
}

/*
  Closes the given iterator.
*/
void
seq_close_iter (seq_iter si)
{
  *si = NULL;
}
