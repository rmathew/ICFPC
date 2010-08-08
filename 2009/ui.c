#include <stdbool.h>
#include <stdio.h>
#include <stdlib.h>
#include <SDL.h>

#include "ovm.h"
#include "ui.h"

#define X_RES 640
#define Y_RES 480

/* Pixels per metre. Determined by the size of the "earth" BMP. */
#define PPM (earth_bmp_width / (2.0 * EARTH_RADIUS))

#define SAT_SIZE 2

static Uint32 earth_bmp_width;

static long double min_x, max_x;
static long double min_y, max_y;

static SDL_Surface *screen = NULL;
static SDL_Surface *earth = NULL;

static Uint32 own_sat_clr;
static Uint32 own_path_clr;

static Uint32 fs_clr;
static Uint32 fs_path_clr;

static Uint32 oth_sat_clr;
static Uint32 oth_path_clr;

static Uint32 bg_clr;

static int num_upd_rects;
static SDL_Rect upd_rects[2 + 2 + 2 * MAX_OTH_SATS];

static char tmp_buf[1024];


/*
static void
draw_pixel (Uint16 x, Uint16 y, Uint32 colour)
{
  Uint8 *bufp_8;
  Uint16 *bufp_16;
  Uint32 *bufp_32;
  Uint8 r, g, b;

  switch (screen->format->BytesPerPixel)
  {
  case 1:
    bufp_8 = (Uint8 *)screen->pixels + y*screen->pitch + x;
    *bufp_8 = colour;
    break;

  case 2:
    bufp_16 = (Uint16 *)screen->pixels + y*screen->pitch/2 + x;
    *bufp_16 = colour;
    break;

  case 3:
    SDL_GetRGB (colour, screen->format, &r, &g, &b);
    bufp_8 = (Uint8 *)screen->pixels + y*screen->pitch + x;
    *(bufp_8 + screen->format->Rshift/8) = r;
    *(bufp_8 + screen->format->Gshift/8) = g;
    *(bufp_8 + screen->format->Bshift/8) = b;
    break;

  case 4:
    bufp_32 = (Uint32 *)screen->pixels + y*screen->pitch/4 + x;
    *bufp_32 = colour;
    break;
  }
}
*/


static bool
is_visible (long double x, long double y)
{
  return (x > min_x) && (x < max_x) && (y > min_y) && (y < max_y);
}


static void
draw_earth (void)
{
  SDL_Rect r;

  r.x = (X_RES - earth_bmp_width) / 2;
  r.y = (Y_RES - earth_bmp_width) / 2;
  r.w = r.h = earth_bmp_width;

  SDL_BlitSurface (earth, NULL, screen, &r);
}


static void
draw_sat (long double x, long double y, Uint32 sat_clr, Uint32 sat_path_clr)
{
  if (is_visible (x, y))
  {
    upd_rects[num_upd_rects].x = upd_rects[num_upd_rects + 1].x;
    upd_rects[num_upd_rects].y = upd_rects[num_upd_rects + 1].y;
    upd_rects[num_upd_rects].w = upd_rects[num_upd_rects + 1].w;
    upd_rects[num_upd_rects].h = upd_rects[num_upd_rects + 1].h;

    upd_rects[num_upd_rects + 1].x = X_RES / 2 + x * PPM;
    upd_rects[num_upd_rects + 1].y = Y_RES / 2 - y * PPM;

    upd_rects[num_upd_rects + 1].w = SAT_SIZE;
    upd_rects[num_upd_rects + 1].h = SAT_SIZE;

    SDL_FillRect (screen, &upd_rects[num_upd_rects], sat_path_clr);
    /*
    draw_pixel
      (upd_rects[num_upd_rects].x, upd_rects[num_upd_rects].y, sat_path_clr);
    */
    SDL_FillRect (screen, &upd_rects[num_upd_rects + 1], sat_clr);

    num_upd_rects += 2;
  }
}


static void
scale_ui (earth_size es)
{
  switch (es)
  {
  case BIG_EARTH:
    earth_bmp_width = 256;
    break;

  case MEDIUM_EARTH:
    earth_bmp_width = 32;
    break;

  case SMALL_EARTH:
    earth_bmp_width = 8;
    break;
  }

  /* Calculate clipping viewport in space for the UI. */
  max_x = (long double) X_RES * EARTH_RADIUS / (long double) earth_bmp_width;
  min_x = -max_x;
  max_y = (long double) Y_RES * EARTH_RADIUS / (long double) earth_bmp_width;
  min_y = -max_y;
}


void
init_ui (char *p_name, uint32_t s_id, earth_size es)
{
  scale_ui (es);

  if (SDL_Init (SDL_INIT_VIDEO) == -1)
  {
    fprintf (stderr, "\nERROR: SDL Initialisation: %s\n", SDL_GetError ());
    exit (1);
  }

  atexit (SDL_Quit);

  snprintf
    (tmp_buf, sizeof (tmp_buf), "earth_%ux%u.bmp", earth_bmp_width,
     earth_bmp_width);

  earth = SDL_LoadBMP (tmp_buf);
  if (earth == NULL)
  {
    fprintf (stderr, "\nERROR: SDL LoadBMP: %s\n", SDL_GetError ());
    exit (1);
  }

  const SDL_VideoInfo *vid_info = SDL_GetVideoInfo ();
  if (vid_info == NULL)
  {
    fprintf (stderr, "\nERROR: SDL VideoInfo: %s\n", SDL_GetError ());
    exit (1);
  }

  screen
    = SDL_SetVideoMode (X_RES, Y_RES, vid_info->vfmt->BitsPerPixel, SDL_SWSURFACE);
  if (screen == NULL)
  {
    fprintf (stderr, "\nERROR: SDL SetVideoMode: %s\n", SDL_GetError ());
    exit (1);
  }

  snprintf
    (tmp_buf, sizeof (tmp_buf), "ICFPC 2009 - [#%u] %s (1 px = %.0Lf m)",
     s_id, p_name, 1.0L/PPM);

  SDL_WM_SetCaption (tmp_buf, NULL);
  SDL_ShowCursor (SDL_ENABLE);

  bg_clr = SDL_MapRGB (screen->format, 0x00, 0x00, 0x00);
  
  if (SDL_MUSTLOCK (screen))
    SDL_LockSurface (screen);

  SDL_FillRect (screen, NULL, bg_clr);
  draw_earth ();

  if (SDL_MUSTLOCK (screen))
    SDL_UnlockSurface (screen);

  SDL_UpdateRect (screen, 0, 0, 0, 0);

  upd_rects[0].x = upd_rects[0].y = 0;
  upd_rects[0].w = upd_rects[0].h = 0;

  upd_rects[1].x = upd_rects[1].y = 0;
  upd_rects[1].w = upd_rects[1].h = 0;

  own_sat_clr = SDL_MapRGB (screen->format, 0xFF, 0xFF, 0xFF);
  own_path_clr = SDL_MapRGB (screen->format, 0x00, 0x40, 0x20);

  fs_clr = SDL_MapRGB (screen->format, 0x00, 0xFF, 0x00);
  fs_path_clr = SDL_MapRGB (screen->format, 0x50, 0x50, 0x50);

  oth_sat_clr = SDL_MapRGB (screen->format, 0xFF, 0x00, 0x00);
  oth_path_clr = SDL_MapRGB (screen->format, 0x35, 0x00, 0x50);
}


bool
update_ui (ui_data *uid)
{
  bool cont_sim = true;

  if (SDL_MUSTLOCK (screen))
    SDL_LockSurface (screen);

  num_upd_rects = 0;
  for (int i = 0; i < uid->num_oth; i++)
  {
    draw_sat (uid->oth_x[i], uid->oth_y[i], oth_sat_clr, oth_path_clr);
  }

  if (uid->have_fs)
  {
    draw_sat (uid->fs_x, uid->fs_y, fs_clr, fs_path_clr);
  }

  draw_sat (uid->x, uid->y, own_sat_clr, own_path_clr);

  if (SDL_MUSTLOCK (screen))
    SDL_UnlockSurface (screen);

  SDL_UpdateRects (screen, num_upd_rects, upd_rects);

  bool should_pause = false;

  SDL_Event evt;
  SDL_PollEvent (&evt);

  switch (evt.type)
  {
  case SDL_QUIT:
    // XXX: Get random QUIT messages from SDL. Bug?
    // cont_sim = false;
    break;

  case SDL_KEYDOWN:
    switch (evt.key.keysym.sym)
    {
    case SDLK_ESCAPE:
      cont_sim = false;
      break;

    case SDLK_SPACE:
      should_pause = true;
      break;

    default:
      break;
    }
    break;
  }


  while (should_pause)
  {
    SDL_WaitEvent (&evt);

    if (evt.type == SDL_QUIT)
    {
      cont_sim = false;
      should_pause = false;
    }
    else if (evt.type == SDL_KEYDOWN)
    {
      should_pause = false;
      cont_sim = (evt.key.keysym.sym != SDLK_ESCAPE);
    }
  };

  return cont_sim;
}


void
quit_ui (void)
{
}
