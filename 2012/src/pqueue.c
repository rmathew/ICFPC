#include <stdint.h>
#include <stdbool.h>
#include <stdio.h>
#include <stdlib.h>

#include "pqueue.h"

#define ELT_AT(x) (*(the_queue->heap + (x)))
#define PARENT(x) (((x) - 1) / 2)
#define LEFT_CHILD(x) (((x) * 2) + 1)
#define RIGHT_CHILD(x) (((x) * 2) + 2)

/*
 * A simple heap-based priority-queue implementation based on the outline
 * given in "Data Structures and Algorithms" by Aho, Hopcroft and Ullman.
 */

pqueue_t*
pq_create(int32_t max_elts, int32_t (*cmp_fn_p)(const void*, const void*)) {
  pqueue_t* ret_val = malloc(sizeof(pqueue_t));
  ret_val->max = max_elts;
  ret_val->last = -1;
  ret_val->compare_p = cmp_fn_p;
  ret_val->heap = (void**)malloc(max_elts * sizeof(void*));
  return ret_val;
}

bool
pq_insert(pqueue_t* the_queue, void* the_elt) {
  bool ret_val = false;
  if ((the_queue != NULL) && (the_queue->last < (the_queue->max - 1))) {
    the_queue->last++;
    ELT_AT(the_queue->last) = the_elt;

    int32_t i = the_queue->last;
    int (*compare)(const void*, const void*) = the_queue->compare_p;
    while ((i > 0) && ((*compare)(ELT_AT(i), ELT_AT(PARENT(i))) < 0)) {
      void* tmp = ELT_AT(i);
      ELT_AT(i) = ELT_AT(PARENT(i));
      ELT_AT(PARENT(i)) = tmp;
      i = PARENT(i);
    }
    ret_val = true;
  }
  return ret_val;
}

bool
pq_is_empty(pqueue_t* the_queue) {
  bool ret_val = true;
  if (the_queue != NULL) {
    ret_val = !(the_queue->last >= 0);
  }
  return ret_val;
}

int32_t
pq_size(pqueue_t* the_queue) {
  return ((the_queue != NULL) ? (the_queue->last + 1) : -1);
}

int32_t
pq_capacity(pqueue_t* the_queue) {
  return ((the_queue != NULL) ? (the_queue->max) : -1);
}

void*
pq_delmin(pqueue_t* the_queue) {
  void* ret_val = NULL;
  if ((the_queue != NULL) && (the_queue->last >= 0)) {
    ret_val = ELT_AT(0);
    ELT_AT(0) = ELT_AT(the_queue->last);
    the_queue->last--;

    if (the_queue->last > 0) {
      int32_t i = 0;
      int (*compare)(const void*, const void*) = the_queue->compare_p;
      while (i <= PARENT(the_queue->last)) {
        int32_t left = LEFT_CHILD(i);
        int32_t right = RIGHT_CHILD(i);
        int32_t smaller;
        if (right > the_queue->last) {
          smaller = left;
        } else {
          if ((*compare)(ELT_AT(left), ELT_AT(right)) < 0) {
            smaller = left;
          } else {
            smaller = right;
          }
        }

        if ((*compare)(ELT_AT(i), ELT_AT(smaller)) > 0) {
          void* tmp = ELT_AT(i);
          ELT_AT(i) = ELT_AT(smaller);
          ELT_AT(smaller) = tmp;
          i = smaller;
        } else {
          break;
        }
        fflush(stdout);
      }
    }
  }
  return ret_val;
}

bool
pq_has_elt(pqueue_t* the_queue, void* the_elt, bool
    (*eq_fn_p)(const void*, const void*)) {
  bool ret_val = false;
  if ((the_queue != NULL) && (the_queue->last >= 0)) {
    // FIXME: Ugly linear search; exploit the partially-ordered property.
    int (*compare)(const void*, const void*) = the_queue->compare_p;
    for (int i = 0; i <= the_queue->last; i++) {
      bool equal = (eq_fn_p != NULL) ? ((*eq_fn_p)(ELT_AT(i), the_elt))
          : ((*compare)(ELT_AT(i), the_elt) == 0);

      if (equal) {
        ret_val = true;
        break;
      }
    }
  }
  return ret_val;
}

void
pq_destroy(pqueue_t* the_queue) {
  if (the_queue != NULL) {
    the_queue->max = -1;
    the_queue->last = -1;
    the_queue->compare_p = NULL;
    free(the_queue->heap);
    free(the_queue);
  }
}
