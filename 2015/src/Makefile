# Makefile for the ICFPC 2015 submission of Ranjit Mathew <rmathew@gmail.com>.
DEL = rm -f

CXX = g++
CXXFLAGS += -std=c++11 -pedantic -Wall -O2 -fomit-frame-pointer -pipe

COMMON_OBJECTS = \
    board.o \
    utils.o \
    jsoncpp.o \

PLAYER = play_icfp2015
PLAYER_OBJECTS = \
    play_main.o \
    $(COMMON_OBJECTS) \

PLAYER_LDFLAGS =
PLAYER_LDLIBS =

VISUALIZER = honey
VISUALIZER_OBJECTS = \
    vis_main.o \
    $(COMMON_OBJECTS) \

VISUALIZER_LDFLAGS =
VISUALIZER_LDLIBS =

PROGRAMS=$(PLAYER) $(VISUALIZER)
OBJECTS = \
    $(PLAYER_OBJECTS) \
    $(VISUALIZER_OBJECTS) \
    $(COMMON_OBJECTS) \

.DEFAULT_GOAL := $(PLAYER)
.PHONY: all clean run

all: $(PLAYER) $(VISUALIZER)

$(PLAYER): $(PLAYER_OBJECTS)
	$(CXX) $(PLAYER_LDFLAGS) $(PLAYER_OBJECTS) $(PLAYER_LDLIBS) \
	    -o $(PLAYER)

$(VISUALIZER): $(VISUALIZER_OBJECTS)
	$(CXX) $(VISUALIZER_LDFLAGS) $(VISUALIZER_OBJECTS) \
	    $(VISUALIZER_LDLIBS) -o $(VISUALIZER)

run: $(VISUALIZER)
	$(VISUALIZER)

clean:
	$(DEL) $(OBJECTS) $(PROGRAMS)
