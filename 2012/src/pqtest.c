#include <stdint.h>
#include <stdbool.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

#include "pqueue.h"

static int32_t
cmp_ui32(const void* e1, const void* e2) {
  return (int32_t)((uint32_t)e1 - (uint32_t)e2);
}

static int
chk_ints(void) {
  uint32_t inp[] = { 23, 2357, 0, 99, 12, };
  uint32_t exp[] = { 0, 12, 23, 99, 2357, };
  int inp_length = sizeof(inp) / sizeof(inp[0]);

  pqueue_t* test_q = pq_create(inp_length, cmp_ui32);
  for (int i = 0; i < inp_length; i++) {
    pq_insert(test_q, (void*)inp[i]);
  }

  if (pq_size(test_q) != inp_length) {
    fprintf(stderr, "ERROR: Queue-size is \"%d\" instead of \"%d\".\n",
        pq_size(test_q), inp_length);
    return 1;
  }

  int i = 0;
  while (!pq_is_empty(test_q)) {
    uint32_t next = (uint32_t)pq_delmin(test_q);
    if (next != exp[i]) {
      fprintf(stderr, "ERROR: Got \"%u\" instead of \"%u\" at index \"%d\".\n",
          next, exp[i], i);
      return 1;
    }
    i++;
  }

  if (i != inp_length) {
    fprintf(stderr, "ERROR: Got \"%d\" elements instead of \"%d\".\n",
        i, inp_length);
    return 1;
  }

  if (pq_size(test_q) != 0) {
    fprintf(stderr, "ERROR: Queue-size is \"%d\" instead of \"0\".\n",
        pq_size(test_q));
    return 1;
  }

  pq_destroy(test_q);

  return 0;
}

static int
chk_strings(void) {
  char* inp[] = { "pakeezah", "shyam", "meena", "ram", };
  char* exp[] = { "meena", "pakeezah", "ram", "shyam", };
  int inp_length = sizeof(inp) / sizeof(inp[0]);

  pqueue_t* test_q = pq_create(inp_length,
      (int (*)(const void *, const void*)) strcmp);
  for (int i = 0; i < inp_length; i++) {
    pq_insert(test_q, (void*)inp[i]);
  }

  if (pq_size(test_q) != inp_length) {
    fprintf(stderr, "ERROR: Queue-size is \"%d\" instead of \"%d\".\n",
        pq_size(test_q), inp_length);
    return 1;
  }

  int i = 0;
  while (!pq_is_empty(test_q)) {
    char* next = (char*)pq_delmin(test_q);
    if (strcmp(next, exp[i]) != 0) {
      fprintf(stderr, "ERROR: Got \"%s\" instead of \"%s\" at index \"%d\".\n",
          next, exp[i], i);
      return 1;
    }
    i++;
  }

  if (i != inp_length) {
    fprintf(stderr, "ERROR: Got \"%d\" elements instead of \"%d\".\n",
        i, inp_length);
    return 1;
  }

  if (pq_size(test_q) != 0) {
    fprintf(stderr, "ERROR: Queue-size is \"%d\" instead of \"0\".\n",
        pq_size(test_q));
    return 1;
  }

  pq_destroy(test_q);

  return 0;
}

int
main(int argc, char* argv[]) {
  int ret_val = chk_ints();
  if (ret_val == 0) {
    ret_val = chk_strings();
  }
  return ret_val;
}
