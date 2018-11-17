#!/bin/sh

set -e
go build main

./halite --replay-directory replays/ -vvv --width 32 --height 32 "go run $GOPATH/src/main/MyBot.go" "go run $GOPATH/src/main/MyBot.go"
