#include <memory>

#include "board.h"
#include "utils.h"

using ::std::unique_ptr;

int main(int argc, char* argv[]) {
    CommonArgs common_args;
    if (!ParseCommonArgs(argc, argv, &common_args)) {
        return 1;
    }
    unique_ptr<Board> board(Board::Create(common_args.input_file_name));
    if (!board) {
        return 2;
    }
    return 0;
}
