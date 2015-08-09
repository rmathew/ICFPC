#include "board.h"

#include <algorithm>

#include "input.h"

using ::std::max;
using ::std::min;
using ::std::string;
using ::std::unique_ptr;
using ::std::vector;

Board::Board(BoardConfig* config)
  : config_(config), current_game_(0), current_unit_index_(-1),
  current_seed_(0ULL), num_remaining_units_(-1) {
    current_unit_location_.x = -1;
    current_unit_location_.y = -1;

    board_cells_.resize(config_->height);
    for (int i = 0; i < config_->height; ++i) {
        board_cells_[i].resize(config_->width, EMPTY_CELL);
    }
}

CellState Board::GetCellState(int x, int y) const {
    if (!IsBoardLocationValid(x, y)) {
        return INVALID_CELL;
    }
    return board_cells_[y][x];
}

bool Board::IsOccupiedByCurrentUnit(int x, int y) const {
    if (!IsBoardLocationValid(x, y)) {
        return false;
    }
    // FIXME: What about rotation?
    int trans_x = x - current_unit_location_.x;
    int trans_y = y - current_unit_location_.y;
    const Unit& current_unit = config_->units[current_unit_index_];
    for (const Location& a_location : current_unit.members) {
        if (a_location.x == trans_x && a_location.y == trans_y) {
            return true;
        }
    }
    return false;
}

bool Board::IsCurrentUnitPivot(int x, int y) const {
    if (!IsBoardLocationValid(x, y)) {
        return false;
    }
    // FIXME: What about rotation?
    int trans_x = x - current_unit_location_.x;
    int trans_y = y - current_unit_location_.y;
    const Unit& current_unit = config_->units[current_unit_index_];
    if (current_unit.pivot.x == trans_x && current_unit.pivot.y == trans_y) {
        return true;
    }
    return false;
}

bool Board::MoveCurrentUnit(MoveCommand cmd) {
    int new_x = current_unit_location_.x;
    int new_y = current_unit_location_.y;
    switch (cmd) {
      case MOVE_E:
        new_x++;
        break;

      case MOVE_W:
        new_x--;
        break;

      case MOVE_SE:
        if ((current_unit_location_.y % 2) == 1) {
            new_x++;
        }
        new_y++;
        break;

      case MOVE_SW:
        if ((current_unit_location_.y % 2) == 0) {
            new_x--;
        }
        new_y++;
        break;
    }
    return PlaceCurrentUnitAt(new_x, new_y, true /* spawn_on_failure */);
}

bool Board::TurnCurrentUnit(TurnCommand cmd) {
    // TODO: Implement this.
    return false;
}

bool Board::IsGameOver() const {
    return num_remaining_units_ <= 0;
}

Board* Board::Create(const string& file_name) {
    unique_ptr<BoardConfig> config(new BoardConfig);
    if (!LoadBoardConfig(file_name, config.get())) {
        return nullptr;
    }
    unique_ptr<Board> board(new Board(config.release()));

    for (const Location& location : board->config_->filled_cells) {
        board->board_cells_[location.y][location.x] = FULL_CELL;
    }
    board->current_seed_ = board->config_->source_seeds[0];
    board->num_remaining_units_ = board->config_->source_length;
    if (!board->SpawnNewUnit()) {
        return nullptr;
    }
    return board.release();
}

bool Board::IsBoardLocationValid(int x, int y) const {
    return x >= 0 && x < config_->width && y >= 0 && y < config_->height;
}

bool Board::IsNewLocationValidForCurrentUnit(int x, int y) const {
    if (!IsBoardLocationValid(x, y)) {
        return false;
    }
    // FIXME: What about rotation?
    const Unit& current_unit = config_->units[current_unit_index_];
    for (const Location& a_location : current_unit.members) {
        if (GetCellState(a_location.x + x, a_location.y + y) != EMPTY_CELL) {
            return false;
        }
    }
    const Location& pivot_loc = current_unit.pivot;
    if (GetCellState(pivot_loc.x + x, pivot_loc.y + y) != EMPTY_CELL) {
        return false;
    }
    return true;
}

bool Board::PlaceCurrentUnitAt(int x, int y, bool spawn_on_failure) {
    if (IsNewLocationValidForCurrentUnit(x, y)) {
        current_unit_location_.x = x;
        current_unit_location_.y = y;
        MarkCurrentUnitCellsVisited();
        return true;
    }
    LockCurrentUnit();
    return spawn_on_failure ? SpawnNewUnit() : false;
}

void Board::LockCurrentUnit() {
    const Unit& current_unit = config_->units[current_unit_index_];
    for (const Location& location : current_unit.members) {
        const int x = current_unit_location_.x + location.x;
        const int y = current_unit_location_.y + location.y;
        if (IsBoardLocationValid(x, y)) {
            board_cells_[y][x] = FULL_CELL;
        }
    }
}

void Board::MarkCurrentUnitCellsVisited() {
    // FIXME: What about rotation?
    const int x = current_unit_location_.x;
    const int y = current_unit_location_.y;
    const Unit& current_unit = config_->units[current_unit_index_];
    for (const Location& a_location : current_unit.members) {
        board_cells_[y + a_location.y][x + a_location.x] = VISITED_CELL;
    }
    const Location& pivot_loc = current_unit.pivot;
    board_cells_[y + pivot_loc.y][x + pivot_loc.x] = VISITED_CELL;
}

void Board::ClearVisitedCells() {
    for (int i = 0; i < config_->width; ++i) {
        for (int j = 0; j < config_->height; ++j) {
            if (board_cells_[j][i] == VISITED_CELL) {
                board_cells_[j][i] = EMPTY_CELL;
            }
        }
    }
}

namespace {

int GetRandFromSeed(uint64_t seed) {
    constexpr uint64_t kBits30To16Mask = 0x000000007FFF0000ULL;
    constexpr int kNumRandBits = 16;
    return (seed & kBits30To16Mask) >> kNumRandBits;
}

uint64_t NextRandSeed(uint64_t seed) {
    constexpr uint64_t kRandMultiplier = 1103515245ULL;
    constexpr uint64_t kRandIncrement = 12345ULL;
    constexpr uint64_t kBits31To0Mask = 0x00000000FFFFFFFFULL;
    return (seed * kRandMultiplier + kRandIncrement) & kBits31To0Mask;
}

void GetNewUnitLocation(const Unit& unit, int board_width, int* x, int* y) {
    int min_x = board_width;
    int max_x = 0;
    for (const Location& a_location : unit.members) {
        min_x = min<int>(min_x, a_location.x);
        max_x = max<int>(max_x, a_location.x);
    }
    const int unit_width = (max_x - min_x + 1);
    *x = (board_width - unit_width) / 2;
    *y = 0;
}

}  // namespace

bool Board::SpawnNewUnit() {
    bool spawned = false;
    while (num_remaining_units_ > 0 && !spawned) {
        current_unit_index_ =
          GetRandFromSeed(current_seed_) % config_->units.size();
        current_seed_ = NextRandSeed(current_seed_);
        num_remaining_units_--;

        ClearVisitedCells();

        const Unit& new_unit = config_->units[current_unit_index_];
        int new_unit_x, new_unit_y;
        GetNewUnitLocation(new_unit, config_->width, &new_unit_x, &new_unit_y);
        spawned = PlaceCurrentUnitAt(new_unit_x, new_unit_y,
          false /* spawn_on_failure */);
    }
    return spawned;
}
