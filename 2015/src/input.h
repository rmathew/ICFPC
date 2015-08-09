#ifndef INPUT_H_INCLUDED
#define INPUT_H_INCLUDED

#include <string>

#include "board.h"

bool LoadBoardConfig(const ::std::string& file_name, BoardConfig* config);

#endif /* INPUT_H_INCLUDED */
