#include <stdint.h>
#include <stdbool.h>
#include <stdio.h>

#include "mine.h"

int
main(int argc, char* argv[]) {
  int ret_status = mine_init("<STDIN>", stdin);
  if (ret_status != 0) {
    return 1;
  }

  char cmds[] = "LDRDDUULLLDDLA";
  for(int i = 0; i < sizeof(cmds) && get_status() == PLAYING; i++) {
    putchar(cmds[i]);
  }

  mine_quit();
  return 0;
}
