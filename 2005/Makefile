#
# NOTE: Use GNU make to process this makefile and GCC to compile
# the programs.
#
CXX=g++ -O2 -fomit-frame-pointer -pipe -Wall

COMMON_OBJECTS=\
  bot.o \
  graph.o \
  world.o \

COP_OBJECTS=\
  copmain.o \
  cop.o \
  randcop.o \

ROBBER_OBJECTS=\
  robmain.o \
  robber.o \
  randrob.o \
  justbank.o \

.PHONY: all clean deps

all: cop cop2 robber robber2

realcop: $(COP_OBJECTS) $(COMMON_OBJECTS)
	$(CXX) -o $@ $^

realrobber: $(ROBBER_OBJECTS) $(COMMON_OBJECTS)
	$(CXX) -o $@ $^

cop: realcop
	echo "#!/bin/sh" >cop
	echo ./realcop >>cop
	chmod uog+x cop

cop2: realcop
	echo "#!/bin/sh" >cop2
	echo ./realcop -random >>cop2
	chmod uog+x cop2

robber: realrobber
	echo "#!/bin/sh" >robber
	echo ./realrobber >>robber
	chmod uog+x robber

robber2: realrobber
	echo "#!/bin/sh" >robber2
	echo ./realrobber -random >>robber2
	chmod uog+x robber2

clean:
	rm -f $(COMMON_OBJECTS)
	rm -f $(COP_OBJECTS)
	rm -f $(ROBBER_OBJECTS)
	rm -f realcop realrobber
	rm -f cop robber
	rm -f cop2 robber2

deps: 
	makedepend *.cpp *.hpp

# DO NOT DELETE

bot.o: graph.hpp bot.hpp world.hpp
cop.o: cop.hpp graph.hpp bot.hpp world.hpp
copmain.o: cop.hpp graph.hpp bot.hpp world.hpp
graph.o: graph.hpp bot.hpp world.hpp
justbank.o: graph.hpp world.hpp robber.hpp bot.hpp
randcop.o: graph.hpp cop.hpp bot.hpp world.hpp
randrob.o: graph.hpp robber.hpp bot.hpp world.hpp
robber.o: graph.hpp bot.hpp world.hpp robber.hpp
robmain.o: robber.hpp graph.hpp bot.hpp world.hpp
world.o: world.hpp
bot.o: graph.hpp world.hpp
cop.o: graph.hpp bot.hpp world.hpp
robber.o: graph.hpp bot.hpp world.hpp
