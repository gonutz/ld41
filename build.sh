#!/bin/bash

# build the tools that are necessary to build the final game
go build ./vendor/github.com/gonutz/blob/cmd/blob
go build ./vendor/github.com/gonutz/payload/cmd/payload
go build ./vendor/github.com/gonutz/rsrc

# build the resource file with the icon so the Go build tool adds it to the executable
./rsrc -arch 386 -ico icon.ico -o rsrc_386.syso
./rsrc -arch amd64 -ico icon.ico -o rsrc_amd64.syso

# build a 32 bit Linux executable that runs on both 32 and 64 bit Linux
set GOARCH=386
go build -ldflags "-s -w" -o game
./blob -folder=rsc -out=rsc.blob
./payload -data=rsc.blob -exe=game -output="No-Brain Jogging"

# delete the temporary helper files
rm rsc.blob
rm game
rm blob
rm payload
rm rsrc
