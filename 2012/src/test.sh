#!/bin/bash

function chk_prog() {
  make -s $1
  if [ "$?" != "0" ]
  then
    echo ERROR: Could not create program \"$1\".
    exit 1
  fi
}

chk_prog checker

RET_STATUS=0
TEST_NUM=0
function chk_map() {
  TEST_NUM=`expr $TEST_NUM + 1`
  ./checker ../task/maps/$1.map $2 $3 $4
  if [ "$?" == "0" ]
  then
    echo Test $TEST_NUM \($1\): PASSED
  else
    echo Test $TEST_NUM \($1\): FAILED
    RET_STATUS=1
  fi
}

chk_map contest1 LDRDDUULLLDDL WON 212
chk_map contest1 DD LOST -2
chk_map contest1 DLLLDD ABORTED 94
chk_map contest2 RRUDRRULURULLLLDDDL WON 281
chk_map contest2 RRRRUD LOST 19
chk_map contest10 UUUUULLLLLLLLLLDDLLLLUUULLLLDDLLLLLDRRRRRUUULLLLLLLLUURRRRUUUUULLUULLUURRRLUUURDDRURRRRRUURRRRRRRRRRRRRRRRRDDDDDLLUUULLLLULLLDDDRRUDLLLDDDDDLLDDURRDLDURUURRDDRRRRRUURRDRLULLDDDDLLLDDDDURRDLDDRRRRRDDD WON 3626
chk_map contest10 UUUUUUUUUUULRDD LOST 110

exit "$RET_STATUS"
