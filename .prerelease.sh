#!/bin/sh
cd casket
go get "github.com/tmpim/casket@$1"
go mod tidy
cd ..
