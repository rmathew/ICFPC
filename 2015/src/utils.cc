#include "utils.h"

#include <string.h>

#include <iostream>

using ::std::cerr;
using ::std::endl;

bool ParseCommonArgs(int argc, char* argv[], CommonArgs* common_args) {
    bool next_arg_input_file_name = false;
    bool next_arg_phrase_of_power = false;
    for (int i = 0; i < argc; ++i) {
        if (strcmp(argv[i], "-f") == 0) {
            next_arg_input_file_name = true;
        } else if (strcmp(argv[i], "-p") == 0) {
            next_arg_phrase_of_power = true;
        } else if (next_arg_input_file_name) {
            common_args->input_file_name.assign(argv[i]);
            next_arg_input_file_name = false;
        } else if (next_arg_phrase_of_power) {
            common_args->phrase_of_power.assign(argv[i]);
            next_arg_phrase_of_power = false;
        }
    }
    if (common_args->input_file_name.empty()) {
        cerr << "ERROR: No input-file specified." << endl;
        return false;
    }
    return true;
}
