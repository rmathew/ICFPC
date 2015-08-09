#ifndef BOARD_H_INCLUDED
#define BOARD_H_INCLUDED

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

struct Location {
    int x;
    int y;
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

    CellState GetCellState(int x, int y) const;

    bool IsOccupiedByCurrentUnit(int x, int y) const;

    bool IsCurrentUnitPivot(int x, int y) const;

    bool MoveCurrentUnit(MoveCommand cmd);

    bool TurnCurrentUnit(TurnCommand cmd);

    bool IsGameOver() const;

  private:
    explicit Board(BoardConfig* config);

    bool IsBoardLocationValid(int x, int y) const;

    bool IsUnitLocationValid(const Unit* unit) const;

    bool ReplaceCurrentUnit(Unit* new_unit, bool spawn_on_failure);

    void LockCurrentUnit();

    void MarkCurrentUnitCellsVisited();

    void ClearVisitedCells();

    bool SpawnNewUnit();

    ::std::unique_ptr<BoardConfig> config_;

    ::std::vector<::std::vector<CellState>> board_cells_;
    int current_game_;
    ::std::unique_ptr<Unit> current_unit_;
    uint64_t current_seed_;
    int num_remaining_units_;
};

#endif /* BOARD_H_INCLUDED */
