OBJECTS=um.o
PROG=um
RM=rm -f

CC=gcc

CFLAGS=-pedantic -ansi -fstrict-aliasing -Wall -Wextra -funit-at-a-time \
  -fomit-frame-pointer -O3 -ggdb

$(PROG): $(OBJECTS)
	$(CC) $(CFLAGS) -o $(PROG) $(OBJECTS)

.PHONY: clean
clean:
	$(RM) $(OBJECTS) $(PROG)
