#include <SDL.h>
#include <stdio.h>
#include <stdlib.h>

#include <algorithm>
#include <iostream>
#include <memory>
#include <string>

#include "board.h"
#include "vis_utils.h"
#include "utils.h"

using ::std::cerr;
using ::std::cout;
using ::std::endl;
using ::std::max;
using ::std::min;
using ::std::string;
using ::std::unique_ptr;

namespace {

constexpr Uint16 kScreenWidth = 1024;
constexpr Uint16 kScreenHeight = 768;
constexpr Uint16 kCellPadding = 1;
// Includes 2x inter-cell padding.
constexpr Uint16 kMinCellWidth = 8;
constexpr Uint16 kPivotWidth = 4;

struct ScreenInfo {
    // Not owned.
    SDL_Surface* screen;
    Uint32 empty_cell_color;
    Uint32 visited_cell_color;
    Uint32 full_cell_color;
    Uint32 unit_cell_color;
    Uint32 pivot_color;
    Uint32 background_color;
    // Includes 2x inter-cell padding.
    Uint16 cell_width;
};

class ScopedSurfaceLocker {
  public:
    explicit ScopedSurfaceLocker(SDL_Surface* surface) : surface_(surface) {
        if (SDL_MUSTLOCK(surface_)) {
            SDL_LockSurface(surface_);
        }
    }

    ~ScopedSurfaceLocker() {
        if (SDL_MUSTLOCK(surface_)) {
            SDL_UnlockSurface(surface_);
        }
    }

  private:
    // Not owned.
    SDL_Surface* surface_;
};

bool InitVisualizer(const string& board_id, Uint16 width, Uint16 height,
  bool full_screen, ScreenInfo* screen_info) {
    if (SDL_Init(SDL_INIT_VIDEO | SDL_INIT_TIMER) == -1) {
        cerr << endl << "ERROR: Could not initialize SDL: " << SDL_GetError()
          << endl;
        return false;
    }
    atexit(SDL_Quit);

    // Not owned.
    const SDL_VideoInfo* vid_info = SDL_GetVideoInfo();
    if (vid_info == nullptr) {
        cerr << "ERROR: Could not get video-information: " << SDL_GetError()
          << endl;
        return false;
    }

    Uint32 vid_flags = SDL_SWSURFACE;
    if (full_screen) {
        vid_flags |= SDL_FULLSCREEN;
    }

    // Not owned.
    SDL_Surface* screen = SDL_SetVideoMode(width, height,
      vid_info->vfmt->BitsPerPixel, vid_flags);
    if (screen == nullptr) {
        cerr << "Could not set " << width << "x" << height << ", "
          << vid_info->vfmt->BitsPerPixel << "-bpp video-mode: "
          << SDL_GetError();
        return false;
    }

    char tmp_buf[256];
    snprintf(tmp_buf, 256, "ICFPC 2015 Honeycomb Visualizer (%s)",
      board_id.c_str());
    SDL_WM_SetCaption(tmp_buf, nullptr);

    SDL_ShowCursor(full_screen ? SDL_DISABLE : SDL_ENABLE);
    SDL_EventState(SDL_MOUSEMOTION, SDL_IGNORE);
    SDL_EventState(SDL_MOUSEBUTTONDOWN, SDL_IGNORE);
    SDL_EventState(SDL_MOUSEBUTTONUP, SDL_IGNORE);
    SDL_EnableKeyRepeat(SDL_DEFAULT_REPEAT_DELAY, SDL_DEFAULT_REPEAT_INTERVAL);

    screen_info->screen = screen;
    // Solarized "base01".
    screen_info->empty_cell_color = SDL_MapRGB(screen->format, 88, 110, 117);
    // Solarized "base1".
    screen_info->visited_cell_color = SDL_MapRGB(screen->format, 147, 161, 161);
    // Solarized "yellow".
    screen_info->full_cell_color = SDL_MapRGB(screen->format, 181, 137, 0);
    // Solarized "orange".
    screen_info->unit_cell_color = SDL_MapRGB(screen->format, 203, 75, 22);
    // Solarized "base02".
    screen_info->pivot_color = SDL_MapRGB(screen->format, 7, 54, 66);
    // Solarized "base03".
    screen_info->background_color = SDL_MapRGB(screen->format, 0, 43, 54);

    // Paint the background.
    {
        ScopedSurfaceLocker sl(screen);
        SDL_Rect screen_rect;
        screen_rect.x = 0;
        screen_rect.y = 0;
        screen_rect.w = screen->w;
        screen_rect.h = screen->h;
        SDL_FillRect(screen, &screen_rect, screen_info->background_color);
    }
    SDL_UpdateRect(screen, 0, 0, 0, 0);

    return true;
}

bool IsBoardSizeOk(const Board* board, ScreenInfo* screen_info) {
    const Uint32 avail_screen_width = screen_info->screen->w;
    const BoardConfig* config = board->GetConfig();
    // Odd rows are displaced to the right.
    const Uint32 width_needed =
      kMinCellWidth * config->width + kMinCellWidth / 2;
    if (width_needed >= avail_screen_width) {
        cerr << "ERROR: Board too wide (" << config->width
          << ") to be displayed." << endl;
        return false;
    }

    const Uint32 avail_screen_height = screen_info->screen->h;
    const Uint32 min_cell_height = GetHexHeight(kMinCellWidth);
    // Rows overlap vertically; effective cell-height is 3/4 of cell-height.
    // The last row needs more than the average height.
    const Uint32 height_needed =
      3 * min_cell_height * config->height / 4 + min_cell_height / 4;
    if (height_needed >= avail_screen_height) {
        cerr << "ERROR: Board too tall (" << config->height
          << ") to be displayed." << endl;
        return false;
    }

    // Solve: x * board_width + x/2 = screen_width
    const Uint16 max_cw_by_width =
      2 * avail_screen_width / (2 * config->width + 1);
    // Solve: 3/4 * y * board_height + y/4 = screen_height
    const Uint16 max_ch_by_height =
      4 * avail_screen_height / (3 * config->height + 1);
    const Uint16 max_cw_by_height = GetHexWidth(max_ch_by_height);
    screen_info->cell_width = min<Uint16>(max_cw_by_width, max_cw_by_height);

    return true;
}

Uint32 GetCellColor(const Board* board, int x, int y,
  const ScreenInfo* screen_info) {
    const CellState cell_state = board->GetCellState(x, y);
    Uint32 cell_color = 0ULL;
    switch (cell_state) {
      case EMPTY_CELL:
        cell_color = screen_info->empty_cell_color;
        break;

      case VISITED_CELL:
        if (board->IsOccupiedByCurrentUnit(x, y)) {
            cell_color = screen_info->unit_cell_color;
        } else {
            cell_color = screen_info->visited_cell_color;
        }
        break;

      case FULL_CELL:
        cell_color = screen_info->full_cell_color;
        break;

      default:
        cerr << "ERROR: Unexpected cell-state." << endl;
        break;
    }
    return cell_color;
}

void DisplayBoard(const Board* board, const ScreenInfo* screen_info) {
    const Uint16 avg_cell_height =
      3 * GetHexHeight(screen_info->cell_width) / 4;
    const BoardConfig* config = board->GetConfig();
    const Uint16 init_x = (screen_info->screen->w -
      screen_info->cell_width * config->width -
      screen_info->cell_width / 2) / 2;
    const Uint16 init_y = (screen_info->screen->h -
      avg_cell_height * config->height - avg_cell_height / 3) / 2;
    const Uint16 drawn_cell_width = screen_info->cell_width - 2 * kCellPadding;
    {
        ScopedSurfaceLocker sl(screen_info->screen);
        Uint16 y = init_y + kCellPadding;
        for (int i = 0; i < config->height; ++i) {
            Uint16 x = init_x + kCellPadding;
            if ((i % 2) == 1) {
                x += (screen_info->cell_width / 2);
            }
            for (int j = 0; j < config->width; ++j) {
                DrawHex(screen_info->screen, x, y, drawn_cell_width,
                  GetCellColor(board, j, i, screen_info));
                if (board->IsCurrentUnitPivot(j, i)) {
                    DrawPivot(screen_info->screen, x + drawn_cell_width / 2,
                      y + 2 * avg_cell_height / 3, kPivotWidth,
                      screen_info->pivot_color);
                }
                x += screen_info->cell_width;
            }
            y += avg_cell_height;
        }
    }
    SDL_UpdateRect(screen_info->screen, 0, 0, 0, 0);
}

bool ProcessInput(Board* board, bool* board_changed) {
    SDL_Event evt;
    SDL_WaitEvent(&evt);
    switch (evt.type) {
      case SDL_QUIT:
        return false;

      case SDL_KEYDOWN:
        switch (evt.key.keysym.sym) {
          case SDLK_ESCAPE:
            return false;

          case SDLK_q:
            return false;

          case SDLK_LEFT:
            board->MoveCurrentUnit(MOVE_W);
            *board_changed = true;
            break;

          case SDLK_RIGHT:
            board->MoveCurrentUnit(MOVE_E);
            *board_changed = true;
            break;

          case SDLK_a:
            board->MoveCurrentUnit(MOVE_SW);
            *board_changed = true;
            break;

          case SDLK_s:
            board->MoveCurrentUnit(MOVE_SE);
            *board_changed = true;
            break;

          default:
            break;
        }
        break;

      default:
        break;
    }
    return true;
}

}  // namespace

int main(int argc, char* argv[]) {
    CommonArgs common_args;
    if (!ParseCommonArgs(argc, argv, &common_args)) {
        return 1;
    }

    unique_ptr<Board> board(Board::Create(common_args.input_file_name));
    if (!board) {
        return 2;
    }
    const BoardConfig* config = board->GetConfig();
    cout << "INFO: Displaying a " << config->width << "x" << config->height
      << " board (" << config->id << ") with " << config->source_length
      << " units per game for " << config->source_seeds.size() << " games."
      << endl;

    ScreenInfo screen_info;
    if (!InitVisualizer(board->GetConfig()->id, kScreenWidth, kScreenHeight,
        false /* full_screen */, &screen_info)) {
        return 3;
    }
    if (!IsBoardSizeOk(board.get(), &screen_info)) {
        return 4;
    }
    DisplayBoard(board.get(), &screen_info);

    bool keep_going = true;
    do {
        bool board_changed = false;
        keep_going = ProcessInput(board.get(), &board_changed);
        if (board_changed) {
            DisplayBoard(board.get(), &screen_info);
        }
        if (board->IsGameOver()) {
            cout << "INFO: GAME OVER!" << endl;
            keep_going = false;
        }
    } while (keep_going);

    return 0;
}
