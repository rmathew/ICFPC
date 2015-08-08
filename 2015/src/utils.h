#ifndef UTILS_H_INCLUDED
#define UTILS_H_INCLUDED

#include <string>

struct CommonArgs {
    ::std::string input_file_name;
    ::std::string phrase_of_power;
};

bool ParseCommonArgs(int argc, char* argv[], CommonArgs* common_args);

#endif /* UTILS_H_INCLUDED */
