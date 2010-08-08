#ifndef UI_H_INCLUDED
#define UI_H_INCLUDED

typedef struct
{
  long double x, y;

  bool have_fs;
  long double fs_x, fs_y;

  int num_oth;

  long double oth_x[MAX_OTH_SATS];
  long double oth_y[MAX_OTH_SATS];
} ui_data;

typedef enum
{
  BIG_EARTH,
  MEDIUM_EARTH,
  SMALL_EARTH,
} earth_size;

extern void init_ui (char *p_name, uint32_t s_id, earth_size es);
extern bool update_ui (ui_data *uid);
extern void quit_ui (void);

#endif /* UI_H_INCLUDED */
