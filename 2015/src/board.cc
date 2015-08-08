#include "board.h"

#include <iostream>
#include <fstream>

#include "jsoncpp.h"

using ::std::cerr;
using ::std::endl;
using ::std::ifstream;
using ::std::string;
using ::std::vector;

Board::Board(const string& id, int width, int height)
  : id_(id), width_(width), height_(height) {}

namespace {

bool ParseInputFile(const string& file_name, Json::Value* root) {
    ifstream ifs(file_name);
    if (!ifs.is_open()) {
        cerr << "ERROR: Could not open the input file \"" << file_name << "\"."
            << endl;
        return false;
    }

    Json::CharReaderBuilder rbuilder;
    rbuilder["collectComments"] = false;
    string errors;
    bool ok = Json::parseFromStream(rbuilder, ifs, root, &errors);
    ifs.close();
    if (!ok) {
        cerr << "ERROR: Invalid JSON in the input file \""
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
        id->assign(root["id"].asString());
    }
    return true;
}

void ParseCell(const Json::Value& cell_data, Cell* cell) {
    if (cell_data.isNull()) {
        return;
    }
    cell->x = cell_data.get("x", -1).asInt();
    cell->y = cell_data.get("y", -1).asInt();
}

void ParseCellArray(const Json::Value& cell_array_data,
  vector<Cell>* cell_array) {
    if (cell_array_data.isNull()) {
        return;
    }
    cell_array->resize(cell_array_data.size());
    for (Json::ArrayIndex i = 0; i < cell_array_data.size(); ++i) {
        ParseCell(cell_array_data[i], &cell_array->at(i));
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
        ParseCellArray(a_unit_data["members"], &a_unit->members);
        ParseCell(a_unit_data["pivot"], &a_unit->pivot);
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

void PopulateBoard(const Json::Value& root_data, vector<Unit>* units,
  vector<Cell>* filled_cells, int* source_length, vector<int>* source_seeds) {
    ParseUnits(root_data["units"], units);
    ParseCellArray(root_data["filled"], filled_cells);
    *source_length = root_data.get("sourceLength", -1).asInt();
    ParseIntArray(root_data["sourceSeeds"], source_seeds);
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

    Board* board = new Board(id, width, height);
    PopulateBoard(root, &board->units_, &board->filled_cells_,
      &board->source_length_, &board->source_seeds_);
    return board;
}
