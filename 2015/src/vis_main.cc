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

constexpr Uint16 kScreenWidth = 800;
constexpr Uint16 kScreenHeight = 600;
constexpr Uint16 kMinCellWidth = 10;
constexpr Uint16 kInterCellPadding = 3;

struct ScreenInfo {
    // Not owned.
    SDL_Surface* screen;
    Uint32 empty_cell_color;
    Uint32 full_cell_color;
    Uint32 unit_cell_color;
    Uint32 background_color;
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

bool IsBoardSizeOk(const Board* board, ScreenInfo* screen_info) {
    constexpr Uint32 kEffectiveMinCellWidth = kMinCellWidth + kInterCellPadding;
    // Odd rows are displaced to the right.
    const Uint32 width_needed =
      kEffectiveMinCellWidth * board->GetWidth() + kEffectiveMinCellWidth / 2;
    if (width_needed >= kScreenWidth) {
        cerr << "ERROR: Board too wide to be displayed." << endl;
        return false;
    }

    const Uint32 min_cell_height = GetHexHeight(kMinCellWidth);

    // Rows overlap vertically.
    const Uint32 effective_min_cell_height =
      (3 * min_cell_height / 4) + kInterCellPadding;
    // The last row needs more than the average space.
    const Uint32 height_needed =
      effective_min_cell_height * board->GetHeight() + min_cell_height / 4 +
      kInterCellPadding;
    if (height_needed >= kScreenHeight) {
        cerr << "ERROR: Board too tall to be displayed." << endl;
        return false;
    }

    // Solve: (x + padding) * board_width + (x + padding) / 2 = screen_width
    const Uint16 max_cw_by_width
      = (2 * kScreenWidth / (2 * board->GetWidth() + 1)) - kInterCellPadding;
    // Solve: (y + padding) * board_height + (y / 3 + padding) = screen_height
    const Uint16 max_avg_height =
      ((3 * kScreenHeight - 2 * kInterCellPadding) /
       (3 * board->GetHeight() + 1)) - kInterCellPadding;
    const Uint16 max_cw_by_height = GetHexWidth(4 * max_avg_height / 3);
    screen_info->cell_width = min<Uint16>(max_cw_by_width, max_cw_by_height);

    return true;
}

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
    // Solarized "yellow".
    screen_info->full_cell_color = SDL_MapRGB(screen->format, 181, 137, 0);
    // Solarized "blue".
    screen_info->unit_cell_color = SDL_MapRGB(screen->format, 38, 139, 210);
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

void DisplayBoard(Board* board, ScreenInfo* screen_info) {
    const Uint16 cell_height = GetHexHeight(screen_info->cell_width);
    const Uint16 eff_cell_width = screen_info->cell_width + kInterCellPadding;
    const Uint16 eff_cell_height = (3 * cell_height / 4) + kInterCellPadding;
    const Uint16 init_x = (kScreenWidth - eff_cell_width * board->GetWidth() -
      screen_info->cell_width / 2) / 2;
    Uint16 y = (kScreenHeight - eff_cell_height * board->GetHeight() -
      cell_height / 4 - kInterCellPadding) / 2;
    {
        ScopedSurfaceLocker sl(screen_info->screen);
        for (int i = 0; i < board->GetHeight(); ++i) {
            Uint16 x = init_x;
            if ((i % 2) == 1) {
                x += (screen_info->cell_width / 2);
            }
            for (int j = 0; j < board->GetWidth(); ++j) {
                const CellState cell_state = board->GetCellState(j, i);
                DrawHex(screen_info->screen, x, y, screen_info->cell_width,
                  (cell_state == EMPTY_CELL) ? screen_info->empty_cell_color :
                  screen_info->full_cell_color);
                x += eff_cell_width;
            }
            y += eff_cell_height;
        }
    }
    SDL_UpdateRect(screen_info->screen, 0, 0, 0, 0);
}

bool GetInput() {
    SDL_Event evt;
    SDL_WaitEvent(&evt);
    switch (evt.type) {
      case SDL_QUIT:
        return false;

      case SDL_KEYDOWN:
        switch (evt.key.keysym.sym) {
          case SDLK_ESCAPE:
            return false;

          default:
            break;
        }
        break;

      default:
        break;
    }
    return true;
}

bool UpdateDisplay(Board* board, ScreenInfo* screen_info) {
    DisplayBoard(board, screen_info);
    bool keep_going = GetInput();
    return keep_going;
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
    ScreenInfo screen_info;
    if (!IsBoardSizeOk(board.get(), &screen_info)) {
        return 3;
    }
    cout << "INFO: Displaying a " << board->GetWidth() << "x"
      << board->GetHeight() << " board (" << board->GetId() << ")." << endl;

    if (!InitVisualizer(board->GetId(), kScreenWidth, kScreenHeight, false,
        &screen_info)) {
        return 4;
    }
    bool keep_going = true;
    do {
        keep_going = UpdateDisplay(board.get(), &screen_info);
    } while (keep_going);

    return 0;
}
