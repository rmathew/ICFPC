#include <stdint.h>
#include <stdbool.h>
#include <stdio.h>
#include <SDL.h>

#include "mine.h"
#include "sdltxt.h"

#define MAX_CMDS 2048U

#define MAX_STATUS_CHARS 31
#define MAX_SCORE_CHARS 63

#define X_MARGIN 5
#define Y_MARGIN 5

#define ICON_SIZE 16U

static char cmds[MAX_CMDS];
static uint32_t num_cmds = 0U;

static char score_board[MAX_SCORE_CHARS + 1];
static char status_board[MAX_STATUS_CHARS + 1];

static char* statuses[] = {
  "PLAYING", "WON", "LOST", "ABORTED",
};

static char* icon_bmps[] = {
  "lambda", "miner", "openlift", "lift", "rock", "bricks", "earth",
};

static SDL_Surface* icons[] = {
  NULL, NULL, NULL, NULL, NULL, NULL, NULL,
};

static Uint16 scr_width = 640;
static Uint16 scr_height = 480;
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

  char tmp_buf[1024];
  sprintf(tmp_buf, "ICFPC 2012 Mine Visualizer (%s)", get_mine_name());
  SDL_WM_SetCaption(tmp_buf, NULL);

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
      entity_t cell_type = get_entity_at(j, i);
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

  dest.x = X_MARGIN;
  dest.y = (scr_height - Y_MARGIN - font_height);
  dest.w = scr_width - (2 * X_MARGIN);
  dest.h = font_height;
  SDL_FillRect(screen, &dest, bg_clr);

  int32_t the_score = get_score();
  sprintf(score_board, "Score: %d", the_score);
  dest.x = X_MARGIN;
  dest.y = (scr_height - Y_MARGIN - font_height);
  sdltxt_write(score_board, MAX_SCORE_CHARS, screen, dest.x, dest.y);

  robot_cond_t cond = get_robot_condition();
  sprintf(status_board, "Status: %s", statuses[cond]);
  dest.x = (scr_width - X_MARGIN - (strlen(status_board) * font_width));
  dest.y = (scr_height - Y_MARGIN - font_height);
  sdltxt_write(status_board, MAX_STATUS_CHARS, screen, dest.x, dest.y);

  dest.x = X_MARGIN;
  dest.y = (scr_height - Y_MARGIN - font_height);
  dest.w = scr_width - (2 * X_MARGIN);
  dest.h = font_height;
  SDL_UpdateRects(screen, 1, &dest);

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
  char cmd_ltr = '\0';
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
      cmd_ltr = 'U';
      break;
    case SDLK_DOWN:
      the_cmd = MOVE_DOWN;
      cmd_ltr = 'D';
      break;
    case SDLK_LEFT:
      the_cmd = MOVE_LEFT;
      cmd_ltr = 'L';
      break;
    case SDLK_RIGHT:
      the_cmd = MOVE_RIGHT;
      cmd_ltr = 'R';
      break;
    case SDLK_a:
      the_cmd = ABORT;
      cmd_ltr = 'A';
      break;
    case SDLK_ESCAPE:
      go_on = false;
      break;
    case SDLK_w:
      the_cmd = WAIT;
      cmd_ltr = 'W';
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

  if (go_on && (the_cmd != UNKNOWN) && (get_robot_condition() == PLAYING)) {
    cmds[num_cmds++] = cmd_ltr;
    refresh_mine(the_cmd);
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

    cmds[num_cmds] = '\0';
    printf("\nExecuted Commands: %s\n", cmds);

    printf("Final State: %s\n", statuses[get_robot_condition()]);

    printf("Final Score: %d\n", get_score());

  }

  if (ret_status == 0) {
    ret_status = vis_quit();
    mine_quit();
  }

  return ret_status;
}
