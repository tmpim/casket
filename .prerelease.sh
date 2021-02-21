#!/bin/sh
cd casket
go get "github.com/tmpim/casket@master"
go mod tidy
cd ..
