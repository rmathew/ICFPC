#ifndef PQUEUE_H_INCLUDED
#define PQUEUE_H_INCLUDED

typedef struct {
  int32_t max;
  int32_t last;
  int (*compare_p)(const void*, const void*);
  void** heap;
} pqueue_t;

extern pqueue_t* pq_create(int32_t max_elts,
    int32_t (*cost_fn_p)(const void*, const void*));
extern bool pq_is_empty(pqueue_t* the_queue);
extern int32_t pq_size(pqueue_t* the_queue);
extern int32_t pq_capacity(pqueue_t* the_queue);
extern bool pq_insert(pqueue_t* the_queue, void* the_elt);
extern void* pq_delmin(pqueue_t* the_queue);
extern void pq_destroy(pqueue_t* the_queue);

#endif /* PQUEUE_H_INCLUDED */
