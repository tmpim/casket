#!/bin/sh
sleep 20 # allow github to catch up
cd casket
GOPROXY=direct go get "github.com/tmpim/casket@master"
go mod tidy
cd ..
