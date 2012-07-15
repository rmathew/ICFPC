#!/bin/bash

if [ -z "$1" ]
then
  echo ERROR: Missing path to test map.
  exit 1
fi

if [ ! -r "$1" ]
then
  echo ERROR: Can not read map file \"$1\".
  exit 1
fi

function chk_prog() {
  make $1
  if [ "$?" != "0" ]
  then
    echo ERROR: Could not create program \"$1\".
    exit 1
  fi
}

chk_prog lifter
chk_prog vis

cat "$1" | ./lifter | ./vis -p "$1"
