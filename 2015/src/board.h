#ifndef BOARD_H_INCLUDED
#define BOARD_H_INCLUDED

#include <cstddef>
#include <cstdint>
#include <memory>
#include <string>
#include <vector>

enum CellState {
    INVALID_CELL,
    EMPTY_CELL,
    VISITED_CELL,
    FULL_CELL,
};

enum MoveCommand {
    MOVE_E,
    MOVE_W,
    MOVE_SE,
    MOVE_SW,
};

enum TurnCommand {
    TURN_CLOCKWISE,
    TURN_COUNTER_CLOCKWISE,
};

// Offset-based location used by the task.
struct Location {
    int row;
    int col;
};

// "Cube"-based location used by http://www.redblobgames.com/grids/hexagons/.
struct CubeLocation {
    int x;
    int y;
    int z;
};

struct Unit {
    ::std::vector<Location> members;
    Location pivot;
};

struct BoardConfig {
    ::std::string id;
    ::std::vector<Unit> units;
    int width;
    int height;
    ::std::vector<Location> filled_cells;
    int source_length;
    ::std::vector<int> source_seeds;
};

class Board {
  public:
    static Board* Create(const ::std::string& file_name);

    const BoardConfig* GetConfig() const {
        return config_.get();
    }

    CellState GetCellState(int row, int col) const;

    bool IsExhausted() {
        return board_exhausted_;
    }

    bool IsOccupiedByCurrentUnit(int row, int col) const;

    bool IsCurrentUnitPivot(int row, int col) const;

    bool MoveCurrentUnit(MoveCommand cmd);

    bool TurnCurrentUnit(TurnCommand cmd);

  private:
    explicit Board(BoardConfig* config);

    bool StartNewGame();

    bool IsLocationValid(int row, int col) const;

    bool IsNewUnitLocationValid(const Unit* new_unit) const;

    bool ReplaceCurrentUnit(Unit* new_unit, bool spawn_on_failure);

    int GetLastFullRow();

    void ClearRow(int row_idx);

    void LockCurrentUnit();

    void MarkCurrentUnitCellsVisited();

    void ClearVisitedCells();

    bool SpawnNewUnit();

    ::std::unique_ptr<BoardConfig> config_;

    ::std::vector<::std::vector<CellState>> board_cells_;
    size_t current_game_;
    ::std::unique_ptr<Unit> current_unit_;
    uint64_t current_seed_;
    int num_remaining_units_;
    bool board_exhausted_;
};

void ToCubeLocation(const Location& loc, CubeLocation* cube_loc);

void FromCubeLocation(const CubeLocation& cube_loc, Location* loc);

#endif /* BOARD_H_INCLUDED */
