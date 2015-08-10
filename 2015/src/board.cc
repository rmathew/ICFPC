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
  num_remaining_units_(-1), board_exhausted_(false) {
    board_cells_.resize(config_->height);
    for (int row = 0; row < config_->height; ++row) {
        board_cells_[row].resize(config_->width, EMPTY_CELL);
    }
}

bool Board::StartNewGame() {
    if (current_game_ >= config_->source_seeds.size()) {
        board_exhausted_ = true;
        return false;
    }
    for (int row = 0; row < config_->height; ++row) {
        board_cells_[row].assign(config_->width, EMPTY_CELL);
    }
    for (const Location& a_location : config_->filled_cells) {
        board_cells_[a_location.row][a_location.col] = FULL_CELL;
    }
    current_seed_ = config_->source_seeds[current_game_];
    num_remaining_units_ = config_->source_length;
    ++current_game_;

    return SpawnNewUnit();
}

CellState Board::GetCellState(int row, int col) const {
    if (!IsLocationValid(row, col)) {
        return INVALID_CELL;
    }
    return board_cells_[row][col];
}

bool Board::IsOccupiedByCurrentUnit(int row, int col) const {
    if (!IsLocationValid(row, col)) {
        return false;
    }
    for (const Location& a_location : current_unit_->members) {
        if (a_location.row == row && a_location.col == col) {
            return true;
        }
    }
    return false;
}

bool Board::IsCurrentUnitPivot(int row, int col) const {
    if (!IsLocationValid(row, col)) {
        return false;
    }
    if (current_unit_->pivot.row == row && current_unit_->pivot.col == col) {
        return true;
    }
    return false;
}

namespace {

void MoveLocation(Location* location, MoveCommand cmd) {
    switch (cmd) {
      case MOVE_E:
        location->col += 1;
        break;

      case MOVE_W:
        location->col -= 1;
        break;

      case MOVE_SE:
        if ((location->row % 2) == 1) {
            location->col += 1;
        }
        location->row += 1;
        break;

      case MOVE_SW:
        if ((location->row % 2) == 0) {
            location->col -= 1;
        }
        location->row += 1;
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

namespace {

void TurnLocation(Location* location, const Location& pivot, TurnCommand cmd) {
    CubeLocation cube_pivot;
    ToCubeLocation(pivot, &cube_pivot);

    CubeLocation cube_location;
    ToCubeLocation(*location, &cube_location);
    cube_location.x -= cube_pivot.x;
    cube_location.y -= cube_pivot.y;
    cube_location.z -= cube_pivot.z;

    int tmp;
    switch (cmd) {
      case TURN_CLOCKWISE:
        tmp = cube_location.z;
        cube_location.z = -cube_location.y;
        cube_location.y = -cube_location.x;
        cube_location.x = -tmp;
        break;

      case TURN_COUNTER_CLOCKWISE:
        tmp = cube_location.x;
        cube_location.x = -cube_location.y;
        cube_location.y = -cube_location.z;
        cube_location.z = -tmp;
        break;
    }

    cube_location.x += cube_pivot.x;
    cube_location.y += cube_pivot.y;
    cube_location.z += cube_pivot.z;
    FromCubeLocation(cube_location, location);
}

}  // namespace

bool Board::TurnCurrentUnit(TurnCommand cmd) {
    unique_ptr<Unit> new_unit(new Unit(*current_unit_));
    for (Location& a_location : new_unit->members) {
        TurnLocation(&a_location, new_unit->pivot, cmd);
    }
    return ReplaceCurrentUnit(new_unit.release(), true /* spawn_on_failure */);
    return false;
}

Board* Board::Create(const string& file_name) {
    unique_ptr<BoardConfig> config(new BoardConfig);
    if (!LoadBoardConfig(file_name, config.get())) {
        return nullptr;
    }

    unique_ptr<Board> board(new Board(config.release()));
    if (!board->StartNewGame()) {
        cerr << "ERROR: Could not start the first game." << endl;
        return nullptr;
    }
    return board.release();
}

bool Board::IsLocationValid(int row, int col) const {
    return row >= 0 && row < config_->height &&
      col >= 0 && col < config_->width;
}

bool Board::IsNewUnitLocationValid(const Unit* new_unit) const {
    bool have_current_unit = (current_unit_ != nullptr);
    if (have_current_unit &&
      (new_unit->members.size() != current_unit_->members.size())) {
        return false;
    }
    for (size_t i = 0; i < new_unit->members.size(); ++i) {
        const Location& new_location = new_unit->members.at(i);
        if (have_current_unit) {
            const Location& current_location = current_unit_->members.at(i);
            if (new_location.row == current_location.row &&
              new_location.col == current_location.col) {
                continue;
            }
        }
        if (GetCellState(new_location.row, new_location.col) != EMPTY_CELL) {
            return false;
        }
    }
    return true;
}

bool Board::ReplaceCurrentUnit(Unit* new_unit, bool spawn_on_failure) {
    unique_ptr<Unit> the_unit(new_unit);
    if (IsNewUnitLocationValid(the_unit.get())) {
        current_unit_.reset(the_unit.release());
        MarkCurrentUnitCellsVisited();
        return true;
    }
    LockCurrentUnit();
    return spawn_on_failure ? SpawnNewUnit() : false;
}

int Board::GetLastFullRow() {
    for (int row_idx = config_->height - 1; row_idx >= 0; --row_idx) {
        bool all_full = true;
        for (const CellState& a_cell_state : board_cells_[row_idx]) {
            if (a_cell_state != FULL_CELL) {
                all_full = false;
                break;
            }
        }
        if (all_full) {
            return row_idx;
        }
    }
    return -1;
}

void Board::ClearRow(int row_idx) {
    for (int i = row_idx; i >= 0; --i) {
        if (i > 0) {
            board_cells_[i] = board_cells_[i - 1];
        } else {
            board_cells_[0].assign(config_->width, EMPTY_CELL);
        }
    }
}

void Board::LockCurrentUnit() {
    for (const Location& a_location : current_unit_->members) {
        if (IsLocationValid(a_location.row, a_location.col)) {
            board_cells_[a_location.row][a_location.col] = FULL_CELL;
        }
    }
    int last_full_row = GetLastFullRow();
    while (last_full_row != -1) {
        ClearRow(last_full_row);
        last_full_row = GetLastFullRow();
    }
}

void Board::MarkCurrentUnitCellsVisited() {
    for (const Location& a_location : current_unit_->members) {
        if (IsLocationValid(a_location.row, a_location.col)) {
            board_cells_[a_location.row][a_location.col] = VISITED_CELL;
        }
    }
}

void Board::ClearVisitedCells() {
    for (int row = 0; row < config_->height; ++row) {
        for (int col = 0; col < config_->width; ++col) {
            if (board_cells_[row][col] == VISITED_CELL) {
                board_cells_[row][col] = EMPTY_CELL;
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
    int min_col = board_width;
    int max_col = 0;
    for (const Location& a_location : unit->members) {
        min_col = min<int>(min_col, a_location.col);
        max_col = max<int>(max_col, a_location.col);
    }
    const int unit_width = (max_col - min_col + 1);
    const int delta_col = (board_width - unit_width) / 2;
    for (Location& a_location : unit->members) {
        a_location.col += delta_col;
    }
    unit->pivot.col += delta_col;
}

}  // namespace

bool Board::SpawnNewUnit() {
    if (num_remaining_units_ <= 0) {
        return false;
    }
    int new_unit_index =
      GetRandFromSeed(current_seed_) % config_->units.size();
    current_seed_ = NextRandSeed(current_seed_);
    num_remaining_units_--;

    ClearVisitedCells();

    unique_ptr<Unit> new_unit(new Unit(config_->units[new_unit_index]));
    SetNewUnitLocation(new_unit.get(), config_->width);
    bool spawn_ok = ReplaceCurrentUnit(new_unit.release(),
      false /* spawn_on_failure */);

    if (!spawn_ok) {
        return StartNewGame();
    }
    return true;
}

void ToCubeLocation(const Location& loc, CubeLocation* cube_loc) {
    cube_loc->x = loc.col - (loc.row - (loc.row & 1)) / 2;
    cube_loc->z = loc.row;
    cube_loc->y = -cube_loc->x - cube_loc->z;
}

void FromCubeLocation(const CubeLocation& cube_loc, Location* loc) {
    loc->col = cube_loc.x + (cube_loc.z - (cube_loc.z & 1)) / 2;
    loc->row = cube_loc.z;
}
