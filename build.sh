#!/bin/bash
set GOARCH=386
go build -ldflags "-s -w" -o game
blob -folder=rsc -out=rsc.blob
payload -data=rsc.blob -exe=game -output="No-Brain Jogging"
rm game
