CC=gcc

GC_PREFIX=/home/rmathew/tmp/gc6.8
GC_INCLUDE=$(GC_PREFIX)/include
GC_LIB=$(GC_PREFIX)/gc.a

DEBUG_CFLAGS=-g
WARN_FLAGS=-Wall
OPT_CFLAGS=-O3 -fomit-frame-pointer -funit-at-a-time -pipe -march=pentium3 \
  -mtune=pentium3
CFLAGS=$(DEBUG_CFLAGS) $(OPT_CFLAGS) $(WARN_FLAGS) -I$(GC_INCLUDE)

LFLAGS=$(PROFILE_FLAGS)
LIBS=$(GC_LIB)

EXE=exe
EXE_OBJS=\
  seq.o \
  exe.o \

BLD=bld
BLD_OBJS=\
  build.o \


.PHONY: all clean

all: $(EXE) $(BLD)

$(EXE): $(EXE_OBJS)
	$(CC) -o $(EXE) $(EXE_OBJS) $(LFLAGS) $(LIBS)

$(BLD): $(BLD_OBJS)
	$(CC) -o $(BLD) $(BLD_OBJS) $(LFLAGS) $(LIBS)

exe.o: seq.h

seq.o: seq.h

clean:
	rm -f $(EXE_OBJS) $(BLD_OBJS)
	rm -f $(EXE) $(BLD)
