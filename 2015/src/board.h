#ifndef BOARD_H_INCLUDED
#define BOARD_H_INCLUDED

#include <string>
#include <vector>

enum CellState {
    INVALID_CELL,
    EMPTY_CELL,
    FULL_CELL,
};

struct Location {
    int x;
    int y;
};

struct Unit {
    ::std::vector<Location> members;
    Location pivot;
};

class Board {
  public:
    static Board* Create(const ::std::string& file_name);

    ::std::string GetId() const {
        return id_;
    }

    int GetWidth() const {
        return width_;
    }

    int GetHeight() const {
        return height_;
    }

    CellState GetCellState(int x, int y) const;

    bool SetCellState(int x, int y, CellState new_state);

  private:
    Board(const ::std::string& id, int width, int height);

    bool IsValidLocation(int x, int y) const;

    ::std::string id_;
    ::std::vector<Unit> units_;
    int width_;
    int height_;
    int source_length_;
    ::std::vector<int> source_seeds_;

    ::std::vector<::std::vector<CellState>> board_state_;
};

#endif /* BOARD_H_INCLUDED */
