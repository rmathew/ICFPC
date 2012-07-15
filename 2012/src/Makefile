DEL=rm -f

SDL_INC=$(shell sdl-config --cflags)
SDL_LIB=$(shell sdl-config --libs)

CC=gcc
CFLAGS=-std=c99 -pedantic -Wall -g -O2 -fomit-frame-pointer -pipe $(SDL_INC)

LIFTER_LFLAGS=

VIS_LFLAGS=$(SDL_LIB) -lX11

COMMON_OBJECTS=\
   mine.o \

LIFTER_OBJECTS=\
  $(COMMON_OBJECTS) \
  lifter.o \

VIS_OBJECTS=\
  $(COMMON_OBJECTS) \
  vis.o \
  sdltxt.o \

LIFTER_PROG=lifter
VIS_PROG=vis

.PHONY: all clean

all: $(LIFTER_PROG) $(VIS_PROG)

$(LIFTER_PROG): $(LIFTER_OBJECTS)
	$(CC) $(CFLAGS) $(LIFTER_OBJECTS) -o $(LIFTER_PROG) $(LIFTER_LFLAGS)

$(VIS_PROG): $(VIS_OBJECTS)
	$(CC) $(CFLAGS) $(VIS_OBJECTS) -o $(VIS_PROG) $(VIS_LFLAGS)

clean:
	$(DEL) $(LIFTER_OBJECTS)
	$(DEL) $(LIFTER_PROG)
	$(DEL) $(VIS_OBJECTS)
	$(DEL) $(VIS_PROG)

# FIXME: Manual enumeration of dependencies.
mine.o: mine.h

lifter.o: mine.h

vis.o: sdltxt.h mine.h

sdltxt.o: sdltxt.h
