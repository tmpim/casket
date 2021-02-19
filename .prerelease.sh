#!/bin/sh
cd casket
go get "github.com/tmpim/casket@$(git tag --points-at HEAD)"
go mod tidy
cd ..
