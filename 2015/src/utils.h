#ifndef UTILS_H_INCLUDED
#define UTILS_H_INCLUDED

#include <string>

struct CmdLine {
    ::std::string input_file_name;
    ::std::string phrase_of_power;
};

bool ParseCommandLine(int argc, char* argv[], CmdLine* cmd_line);

#endif /* UTILS_H_INCLUDED */
