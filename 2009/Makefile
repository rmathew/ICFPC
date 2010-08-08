#
# Makefile for rmathew's submission for ICFPC 2009.
#

DEL=rm -f
TAR_CJVF=tar --exclude "*.obf" --exclude "*.osf" --exclude "*.pdf" -cjvf

BKP_FILE=$(HOME)/tmp/ovm.tar.bz2

SDL_INC=$(shell sdl-config --cflags)
SDL_LIB=$(shell sdl-config --libs)

CC=gcc
CFLAGS=-std=c99 -pedantic -Wall -Werror -ggdb -O2 -fomit-frame-pointer \
  $(SDL_INC)

LFLAGS=$(SDL_LIB) -lm

OBJECTS=\
  main.o \
  ovm.o \
  ui.o \

PROG=ovm

.PHONY: clean bkp

$(PROG): $(OBJECTS)
	$(CC) $(CFLAGS) $(OBJECTS) -o $(PROG) $(LFLAGS)

clean:
	$(DEL) $(OBJECTS)
	$(DEL) $(PROG)

bkp: clean
	$(TAR_CJVF) $(BKP_FILE) *

main.o: ovm.h

ovm.o: ovm.h

ui.o: ui.h ovm.h
