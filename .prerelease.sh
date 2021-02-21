#!/bin/sh
cd casket
go get "github.com/tmpim/casket@$(git describe)"
go mod tidy
cd ..
