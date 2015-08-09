#include "input.h"

#include <iostream>
#include <fstream>
#include <vector>

#include "jsoncpp.h"

using ::std::string;
using ::std::cerr;
using ::std::endl;
using ::std::ifstream;
using ::std::vector;

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

void GetBoardConfig(const Json::Value& root, BoardConfig* config) {
    if (root.isMember("id")) {
        config->id.assign("Board: ");
        config->id.append(root["id"].asString());
    }
    ParseUnits(root["units"], &config->units);
    config->width = root.get("width", -1).asInt();
    config->height = root.get("height", -1).asInt();
    ParseLocationArray(root["filled"], &config->filled_cells);
    config->source_length = root.get("sourceLength", -1).asInt();
    ParseIntArray(root["sourceSeeds"], &config->source_seeds);
}

bool IsValidBoardConfig(const BoardConfig* config) {
    if (config->units.size() == 0) {
        cerr << "ERROR: No Units defined." << endl;
        return false;
    }
    if (config->width <= 0) {
        cerr << "ERROR: Invalid width (" << config->width << ")." << endl;
        return false;
    }
    if (config->height <= 0) {
        cerr << "ERROR: Invalid height (" << config->height << ")." << endl;
        return false;
    }
    if (config->source_length <= 0) {
        cerr << "ERROR: No Units in the source." << endl;
        return false;
    }
    if (config->source_seeds.size() == 0) {
        cerr << "ERROR: No source-seeds defined." << endl;
        return false;
    }
    return true;
}

}  // namespace

bool LoadBoardConfig(const string& file_name, BoardConfig* config) {
    Json::Value root;
    bool parse_ok = ParseInputFile(file_name, &root);
    if (!parse_ok || root.isNull()) {
        return false;
    }
    config->id.assign(file_name);
    GetBoardConfig(root, config);
    if (!IsValidBoardConfig(config)) {
        return false;
    }
    return true;
}
