#include <memory>

#include "board.h"
#include "utils.h"

using ::std::unique_ptr;

int main(int argc, char* argv[]) {
    CmdLine cmd_line;
    if (!ParseCommandLine(argc, argv, &cmd_line)) {
        return 1;
    }
    unique_ptr<Board> board(Board::Create(cmd_line.input_file_name));
    if (!board) {
        return 2;
    }
    return 0;
}
