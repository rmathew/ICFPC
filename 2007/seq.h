#ifndef SEQ_H
#define SEQ_H

struct cons
{
  void *car;
  struct cons *cdr;
};

typedef struct cons **seq_iter;

typedef struct seq
{
  unsigned int size;
  struct cons *head;
  struct cons *tail;
} *seq;

extern seq seq_new (void);
extern void seq_del (seq);

extern unsigned int seq_size (seq);
extern void *seq_elem_at (seq, unsigned int);
extern void seq_discard (seq, unsigned int);

extern void seq_append (seq, void *);
extern void seq_prepend (seq, void *);

extern seq_iter seq_get_iter (seq);
extern int seq_iter_has_next (seq_iter);
extern void *seq_iter_next (seq_iter);
extern void seq_close_iter (seq_iter);

#endif /* SEQ_H */
