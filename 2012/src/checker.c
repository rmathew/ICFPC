#include <stdint.h>
#include <stdbool.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

#include "mine.h"

static char* statuses[] = {
    "PLAYING", "WON", "LOST", "ABORTED",
};

int
main(int argc, char* argv[]) {
  if (argc != 5) {
    fprintf(stderr, "ERROR: Invalid number of arguments.\n");
    fprintf(stderr,
        "Usage: %s <map-file> <cmds> <expected-status> <expected-score>\n",
        argv[0]);
    return 1;
  }

  FILE* fp = fopen(argv[1], "r");
  if (fp == NULL) {
    fprintf(stderr, "ERROR: Unable to open map file \"%s\".\n", argv[1]);
    return 1;
  }

  int ret_status = 0;
  if (mine_init(fp) == 0) {
    char* cmds = argv[2];
    int cmds_len = strlen(cmds);
    for (int i = 0; (i < cmds_len) && (get_status() == PLAYING); i++) {
      refresh_mine(cmds[i]);
    }

    // Implicit ABORT if end-of-input while playing.
    if (get_status() == PLAYING) {
      refresh_mine(CMD_ABORT);
    }

    char* exp_status = argv[3];
    char* got_status = statuses[get_status()];
    if (strcmp(got_status, exp_status) != 0) {
      fprintf(stderr, "ERROR: Got status \"%s\" instead of \"%s\".\n",
          got_status, exp_status);
      ret_status = 1;
    }

    int exp_score = atoi(argv[4]);
    int got_score = get_score();
    if (got_score != exp_score) {
      fprintf(stderr, "ERROR: Got score \"%d\" instead of \"%d\".\n",
          got_score, exp_score);
      ret_status = 1;
    }
  }
  fclose(fp);
  mine_quit();

  return ret_status;
}
