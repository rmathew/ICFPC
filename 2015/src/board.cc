#include "board.h"

#include <algorithm>
#include <cstdio>
#include <iostream>

#include "input.h"

using ::std::cerr;
using ::std::cout;
using ::std::endl;
using ::std::max;
using ::std::min;
using ::std::string;
using ::std::unique_ptr;
using ::std::vector;

Board::Board(BoardConfig* config)
  : config_(config), current_game_(0), current_seed_(0ULL),
  num_remaining_units_(-1) {
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
    for (const Location& a_location : current_unit_->members) {
        if (a_location.x == x && a_location.y == y) {
            return true;
        }
    }
    return false;
}

bool Board::IsCurrentUnitPivot(int x, int y) const {
    if (!IsBoardLocationValid(x, y)) {
        return false;
    }
    if (current_unit_->pivot.x == x && current_unit_->pivot.y == y) {
        return true;
    }
    return false;
}

namespace {

void MoveLocation(Location* location, MoveCommand cmd) {
    switch (cmd) {
      case MOVE_E:
        location->x += 1;
        break;

      case MOVE_W:
        location->x -= 1;
        break;

      case MOVE_SE:
        if ((location->y % 2) == 1) {
            location->x += 1;
        }
        location->y += 1;
        break;

      case MOVE_SW:
        if ((location->y % 2) == 0) {
            location->x -= 1;
        }
        location->y += 1;
        break;
    }
}

}  // namespace

bool Board::MoveCurrentUnit(MoveCommand cmd) {
    unique_ptr<Unit> new_unit(new Unit(*current_unit_));
    for (Location& a_location : new_unit->members) {
        MoveLocation(&a_location, cmd);
    }
    MoveLocation(&new_unit->pivot, cmd);
    return ReplaceCurrentUnit(new_unit.release(), true /* spawn_on_failure */);
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

    for (const Location& a_location : board->config_->filled_cells) {
        board->board_cells_[a_location.y][a_location.x] = FULL_CELL;
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

bool Board::IsUnitLocationValid(const Unit* unit) const {
    for (const Location& a_location : unit->members) {
        if (GetCellState(a_location.x, a_location.y) != EMPTY_CELL) {
            return false;
        }
    }
    return true;
}

bool Board::ReplaceCurrentUnit(Unit* new_unit, bool spawn_on_failure) {
    unique_ptr<Unit> the_unit(new_unit);
    if (IsUnitLocationValid(the_unit.get())) {
        current_unit_.reset(the_unit.release());
        MarkCurrentUnitCellsVisited();
        return true;
    }
    LockCurrentUnit();
    return spawn_on_failure ? SpawnNewUnit() : false;
}

void Board::LockCurrentUnit() {
    for (const Location& a_location : current_unit_->members) {
        if (IsBoardLocationValid(a_location.x, a_location.y)) {
            board_cells_[a_location.y][a_location.x] = FULL_CELL;
        }
    }
}

void Board::MarkCurrentUnitCellsVisited() {
    for (const Location& a_location : current_unit_->members) {
        if (IsBoardLocationValid(a_location.x, a_location.y)) {
            board_cells_[a_location.y][a_location.x] = VISITED_CELL;
        }
    }
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

void SetNewUnitLocation(Unit* unit, int board_width) {
    int min_x = board_width;
    int max_x = 0;
    for (const Location& a_location : unit->members) {
        min_x = min<int>(min_x, a_location.x);
        max_x = max<int>(max_x, a_location.x);
    }
    const int unit_width = (max_x - min_x + 1);
    const int delta_x = (board_width - unit_width) / 2;
    for (Location& a_location : unit->members) {
        a_location.x += delta_x;
    }
    unit->pivot.x += delta_x;
}

}  // namespace

bool Board::SpawnNewUnit() {
    bool spawned = false;
    while (num_remaining_units_ > 0 && !spawned) {
        int new_unit_index =
          GetRandFromSeed(current_seed_) % config_->units.size();
        current_seed_ = NextRandSeed(current_seed_);
        num_remaining_units_--;

        ClearVisitedCells();

        unique_ptr<Unit> new_unit(new Unit(config_->units[new_unit_index]));
        SetNewUnitLocation(new_unit.get(), config_->width);
        spawned = ReplaceCurrentUnit(new_unit.release(),
          false /* spawn_on_failure */);
    }
    return spawned;
}
