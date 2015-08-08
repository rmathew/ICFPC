#ifndef BOARD_H_INCLUDED
#define BOARD_H_INCLUDED

#include <string>
#include <vector>

struct Cell {
    int x;
    int y;
};

struct Unit {
    ::std::vector<Cell> members;
    Cell pivot;
};

class Board {
  public:
    static Board* Create(const ::std::string& file_name);

  private:
    Board(const ::std::string& id, int width, int height);

    ::std::string id_;
    ::std::vector<Unit> units_;
    int width_;
    int height_;
    ::std::vector<Cell> filled_cells_;
    int source_length_;
    ::std::vector<int> source_seeds_;
};

#endif /* BOARD_H_INCLUDED */
