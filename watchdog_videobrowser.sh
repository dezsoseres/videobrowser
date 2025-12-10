#!/bin/bash

PRG=videobrowser
PRGPATH=/opt/videobrowser

for V in 1 2 3 4 5 6 7 8 9 10 11 
do
  if [ ! "$(pidof $PRG)" ]; then	
    cd $PRGPATH 
    # and call it with full path to be able to see current path from ps
    $PRGPATH/$PRG &
    datum=`date '+20%y-%m-%d %H:%M:%S'`
    echo [$datum] webserver restarted! >> operate.log
  fi
  sleep 5
done



