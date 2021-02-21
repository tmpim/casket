#!/bin/sh
sleep 20 # allow github to catch up
cd casket
GOPRIVATE=github.com/tmpim/casket go get "github.com/tmpim/casket@master"
go mod tidy
cd ..
