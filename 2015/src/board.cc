#include "board.h"

#include <iostream>
#include <fstream>
#include <memory>

#include "jsoncpp.h"

using ::std::cerr;
using ::std::endl;
using ::std::ifstream;
using ::std::string;
using ::std::vector;
using ::std::unique_ptr;

Board::Board(const string& id, int width, int height)
  : id_(id), width_(width), height_(height), current_game_(0),
  current_unit_index_(-1), current_seed_(0ULL) {
    current_unit_location_.x = -1;
    current_unit_location_.y = -1;

    board_state_.resize(height);
    for (int i = 0; i < height; ++i) {
        board_state_[i].resize(width, EMPTY_CELL);
    }
}

CellState Board::GetCellState(int x, int y) const {
    if (!IsValidLocation(x, y)) {
        return INVALID_CELL;
    }
    return board_state_[y][x];
}

bool Board::IsOccupiedByUnit(int x, int y) const {
    if (!IsValidLocation(x, y)) {
        return false;
    }
    // FIXME: What about rotation?
    int trans_x = x - current_unit_location_.x;
    int trans_y = y - current_unit_location_.y;
    const Unit& current_unit = available_units_[current_unit_index_];
    for (const Location& a_location : current_unit.members) {
        if (a_location.x == trans_x && a_location.y == trans_y) {
            return true;
        }
    }
    return false;
}

bool Board::IsUnitPivot(int x, int y) const {
    if (!IsValidLocation(x, y)) {
        return false;
    }
    // FIXME: What about rotation?
    int trans_x = x - current_unit_location_.x;
    int trans_y = y - current_unit_location_.y;
    const Unit& current_unit = available_units_[current_unit_index_];
    if (current_unit.pivot.x == trans_x && current_unit.pivot.y == trans_y) {
        return true;
    }
    return false;
}

namespace {

bool ParseInputFile(const string& file_name, Json::Value* root) {
    ifstream ifs(file_name);
    if (!ifs.is_open()) {
        cerr << "ERROR: Could not open the input-file \"" << file_name << "\"."
          << endl;
        return false;
    }

    Json::CharReaderBuilder rbuilder;
    rbuilder["collectComments"] = false;
    string errors;
    bool ok = Json::parseFromStream(rbuilder, ifs, root, &errors);
    ifs.close();
    if (!ok) {
        cerr << "ERROR: Invalid JSON in the input-file \""
          << file_name << "\"." << endl << errors << endl;
    }
    return ok;
}

bool GetBoardAttrs(const Json::Value& root, string* id, int* width,
  int* height) {
    *width = root.get("width", -1).asInt();
    *height = root.get("height", -1).asInt();
    if (width <= 0 || height <= 0) {
        cerr << "ERROR: Invalid width/height (" << width << ", " << height
          << ")." << endl;
        return false;
    }
    if (root.isMember("id")) {
        id->assign("Board: ");
        id->append(root["id"].asString());
    }
    return true;
}

void ParseLocation(const Json::Value& location_data, Location* location) {
    if (location_data.isNull()) {
        return;
    }
    location->x = location_data.get("x", -1).asInt();
    location->y = location_data.get("y", -1).asInt();
}

void ParseLocationArray(const Json::Value& location_array_data,
  vector<Location>* location_array) {
    if (location_array_data.isNull()) {
        return;
    }
    location_array->resize(location_array_data.size());
    for (Json::ArrayIndex i = 0; i < location_array_data.size(); ++i) {
        ParseLocation(location_array_data[i], &location_array->at(i));
    }
}

void ParseUnits(const Json::Value& units_data, vector<Unit>* units) {
    if (units_data.isNull()) {
        return;
    }
    units->resize(units_data.size());
    for (Json::ArrayIndex i = 0; i < units_data.size(); ++i) {
        Unit* a_unit = &units->at(i);
        const Json::Value a_unit_data = units_data[i];
        ParseLocationArray(a_unit_data["members"], &a_unit->members);
        ParseLocation(a_unit_data["pivot"], &a_unit->pivot);
    }
}

void ParseIntArray(const Json::Value& int_array_data, vector<int>* int_array) {
    if (int_array_data.isNull()) {
        return;
    }
    int_array->resize(int_array_data.size());
    for (Json::ArrayIndex i = 0; i < int_array_data.size(); ++i) {
        int_array->at(i) = int_array_data[i].asInt();
    }
}

bool GetInitBoardConfig(const Json::Value& root_data, vector<Unit>* units,
  vector<Location>* filled_cells, int* source_length,
  vector<int>* source_seeds) {
    ParseUnits(root_data["units"], units);
    if (units->size() == 0) {
        cerr << "ERROR: No Units defined." << endl;
        return false;
    }
    ParseLocationArray(root_data["filled"], filled_cells);
    *source_length = root_data.get("sourceLength", -1).asInt();
    if (*source_length <= 0) {
        cerr << "ERROR: No Units in the source." << endl;
        return false;
    }
    ParseIntArray(root_data["sourceSeeds"], source_seeds);
    if (source_seeds->size() == 0) {
        cerr << "ERROR: No source-seeds defined." << endl;
        return false;
    }
    return true;
}

}  // namespace

Board* Board::Create(const string& file_name) {
    Json::Value root;
    bool parse_ok = ParseInputFile(file_name, &root);
    if (!parse_ok || root.isNull()) {
        return nullptr;
    }

    string id(file_name);
    int width, height;
    bool attrs_ok = GetBoardAttrs(root, &id, &width, &height);
    if (!attrs_ok) {
        return nullptr;
    }
    unique_ptr<Board> board(new Board(id, width, height));

    vector<Location> filled_cells;
    bool cfg_ok = GetInitBoardConfig(root, &board->available_units_,
      &filled_cells, &board->source_length_, &board->source_seeds_);
    if (!cfg_ok) {
        return nullptr;
    }

    for (const Location& location : filled_cells) {
        board->board_state_[location.y][location.x] = FULL_CELL;
    }
    board->current_seed_ = board->source_seeds_[0];
    board->SpawnNextUnit();

    return board.release();
}

bool Board::IsValidLocation(int x, int y) const {
    return x >= 0 && x < width_ && y >= 0 && y < height_;
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

}  // namespace

void Board::SpawnNextUnit() {
    current_unit_index_ =
      GetRandFromSeed(current_seed_) % available_units_.size();
    current_seed_ = NextRandSeed(current_seed_);
    // FIXME: Center the Unit.
    current_unit_location_.x = 0;
    current_unit_location_.y = 0;
}
