#!/bin/bash

PRG=videobrowser

# delete previous
rm $PRG 2>/dev/null

# creating executable binary
CGO_ENABLED=0 go build $PRG.go 

# strip and chk
strip $PRG
ls -l
file $PRG
#ldd $PRG


