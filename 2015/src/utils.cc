#include "utils.h"

#include <string.h>

#include <iostream>

using ::std::cerr;
using ::std::endl;

bool ParseCommandLine(int argc, char* argv[], CmdLine* cmd_line) {
    bool next_arg_input_file_name = false;
    bool next_arg_phrase_of_power = false;
    for (int i = 0; i < argc; ++i) {
        if (strcmp(argv[i], "-f") == 0) {
            next_arg_input_file_name = true;
        } else if (strcmp(argv[i], "-p") == 0) {
        } else if (next_arg_input_file_name) {
            cmd_line->input_file_name.assign(argv[i]);
            next_arg_input_file_name = false;
        } else if (next_arg_phrase_of_power) {
            cmd_line->phrase_of_power.assign(argv[i]);
            next_arg_phrase_of_power = false;
        }
    }
    if (cmd_line->input_file_name.empty()) {
        cerr << "ERROR: No input files specified." << endl;
        return false;
    }
    return true;
}
