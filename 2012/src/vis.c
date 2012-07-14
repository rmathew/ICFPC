#include <stdint.h>
#include <stdbool.h>
#include <stdio.h>
#include <SDL.h>

#include "mine.h"
#include "sdltxt.h"

#define ICON_SIZE 16U

static char* icon_bmps[] = {
  "lambda", "miner", "openlift", "lift", "rock", "bricks", "earth",
};

static SDL_Surface* icons[] = {
  NULL, NULL, NULL, NULL, NULL, NULL, NULL,
};

static Uint16 scr_width = 800;
static Uint16 scr_height = 600;
static Uint8 scr_bypp;
static SDL_Surface *screen = NULL;

static Uint16 font_width;
static Uint16 font_height;

static Uint32 bg_clr;

static int
load_icons(void) {
  int ret_status = 0;
  char tmp_buf[256];
  for (int i = 0; i < sizeof (icon_bmps) / sizeof (char *); i++) {
    sprintf(tmp_buf, "./icons/%s.bmp", icon_bmps[i]);
    icons[i] = SDL_LoadBMP(tmp_buf);
    if (icons[i] == NULL) {
      fprintf(stderr, "Could not load BMP \"%s\": %s\n", tmp_buf,
          SDL_GetError());
      ret_status = 1;
      break;
    }
  }
  return ret_status;
}

static int
vis_init(bool full_screen) {
  if (SDL_Init(SDL_INIT_VIDEO | SDL_INIT_TIMER) == -1) {
    fprintf(stderr, "Could not initialize SDL: %s.\n", SDL_GetError());
    return 1;
  }

  atexit(SDL_Quit);

  const SDL_VideoInfo *vid_info = SDL_GetVideoInfo();
  if (vid_info == NULL) {
    fprintf(stderr, "Could not get video information: %s.\n", SDL_GetError());
    return 1;
  }

  Uint32 vid_flags = SDL_SWSURFACE;
  if (full_screen == true) {
    vid_flags |= SDL_FULLSCREEN;
  }

  screen = SDL_SetVideoMode(scr_width, scr_height,
      vid_info->vfmt->BitsPerPixel, vid_flags);
  if (screen == NULL) {
    fprintf(stderr, "Could not set %ux%u %u-bpp video mode: %s.\n", scr_width,
        scr_height, vid_info->vfmt->BitsPerPixel, SDL_GetError ());
    return 1;
  }

  scr_bypp = screen->format->BytesPerPixel;
  SDL_WM_SetCaption("Mine Visualizer", NULL);

  SDL_ShowCursor(full_screen == true ? SDL_DISABLE : SDL_ENABLE);
  SDL_EventState(SDL_MOUSEMOTION, SDL_IGNORE);
  SDL_EventState(SDL_MOUSEBUTTONDOWN, SDL_IGNORE);
  SDL_EventState(SDL_MOUSEBUTTONUP, SDL_IGNORE);
  SDL_EnableKeyRepeat(SDL_DEFAULT_REPEAT_DELAY, SDL_DEFAULT_REPEAT_INTERVAL);

  sdltxt_init(screen->format);
  sdltxt_metrics(&font_width, &font_height);

  SDL_Color clr;
  clr.r = 0x00;
  clr.g = 0x00;
  clr.b = 0x00;
  bg_clr = SDL_MapRGB(screen->format, clr.r, clr.g, clr.b);

  if (load_icons() != 0) {
    return 1;
  }

  return 0;
}

static int
vis_quit(void) {
  sdltxt_quit();
  return 0;
}

static void
draw_map(void) {
  uint16_t num_rows = get_num_rows();
  uint16_t num_cols = get_num_cols();

  Uint16 x_off = 0U;
  if (scr_width > (num_cols * ICON_SIZE)) {
    x_off = (scr_width - (num_cols * ICON_SIZE)) / 2U;
  }

  Uint16 y_off = 0U;
  if (scr_height > (num_rows * ICON_SIZE)) {
    y_off = (scr_height - (num_rows * ICON_SIZE)) / 2U;
  }

  SDL_Rect dest;
  dest.w = ICON_SIZE;
  dest.h = ICON_SIZE;
  for (int i = 0; i < num_rows; i++) {
    for (int j = 0; j < num_cols; j++) {
      dest.x = x_off + j * ICON_SIZE;
      dest.y = y_off + i * ICON_SIZE;
      entity_t cell_type = get_entity_at(i, j);
      switch (cell_type) {
        case EMPTY_SPACE:
          SDL_FillRect(screen, &dest, bg_clr);
          break;
        default:
          SDL_BlitSurface(icons[cell_type], NULL, screen, &dest);
          break;
      }
    }
  }

  dest.x = x_off;
  dest.y = y_off;
  dest.w = ICON_SIZE * num_cols;
  dest.h = ICON_SIZE * num_rows;
  SDL_UpdateRects(screen, 1, &dest);
}

static bool
vis_update(void) {
  bool go_on = true;

  if (SDL_MUSTLOCK(screen)) {
    if (SDL_LockSurface(screen) < 0) {
      return false;
    }
  }

  draw_map();

  if (SDL_MUSTLOCK(screen)) {
    SDL_UnlockSurface(screen);
  }

  robot_cmd_t the_cmd = UNKNOWN;
  SDL_Event evt;
  SDL_WaitEvent(&evt);
  switch (evt.type) {
  case SDL_QUIT:
    go_on = false;
    break;
  case SDL_KEYDOWN:
    switch (evt.key.keysym.sym) {
    case SDLK_UP:
      the_cmd = MOVE_UP;
      break;
    case SDLK_DOWN:
      the_cmd = MOVE_DOWN;
      break;
    case SDLK_LEFT:
      the_cmd = MOVE_LEFT;
      break;
    case SDLK_RIGHT:
      the_cmd = MOVE_RIGHT;
      break;
    case SDLK_q:
    case SDLK_ESCAPE:
      the_cmd = ABORT;
      break;
    case SDLK_SPACE:
      the_cmd = WAIT;
      break;
    default:
      the_cmd = UNKNOWN;
      break;
    }
    break;
  default:
    the_cmd = UNKNOWN;
    break;
  }

  if (go_on && (the_cmd != UNKNOWN)) {
    update_map(the_cmd);
  }

  if (the_cmd == ABORT) {
    go_on = false;
  }

  return go_on;
}

int
main(int argc, char* argv[]) {
  int ret_status;

  ret_status = mine_init(argc, argv);

  if (ret_status == 0) {
    ret_status = vis_init(false);
  }

  if (ret_status == 0) {
    bool cont_exec;
    do {
      cont_exec = vis_update();
    } while (cont_exec);

    printf("Final Score: %d\n", get_score());
  }

  if (ret_status == 0) {
    ret_status = vis_quit();
  }

  return ret_status;
}
