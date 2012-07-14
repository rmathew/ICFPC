DEL=rm -f

SDL_INC=$(shell sdl-config --cflags)
SDL_LIB=$(shell sdl-config --libs)

CC=gcc
CFLAGS=-std=c99 -pedantic -Wall -g -O2 -fomit-frame-pointer -pipe $(SDL_INC)

VIS_LFLAGS=$(SDL_LIB) -lX11

COMMON_OBJECTS=\
   mine.o \

VIS_OBJECTS=\
  $(COMMON_OBJECTS) \
  vis.o \
  sdltxt.o \

VIS_PROG=vis

.PHONY: clean

$(VIS_PROG): $(VIS_OBJECTS)
	$(CC) $(CFLAGS) $(VIS_OBJECTS) -o $(VIS_PROG) $(VIS_LFLAGS)

clean:
	$(DEL) $(VIS_OBJECTS)
	$(DEL) $(VIS_PROG)

# Manual enumeration of dependencies. FIXME.
mine.o: mine.h

sdltxt.o: sdltxt.h

vis.o: sdltxt.h mine.h
