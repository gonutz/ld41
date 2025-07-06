#!/bin/bash

# Get the necessary build tool.
go install github.com/gonutz/rsrc@latest

REM Build the resource files with the icon so the Go build tool adds it to the executable.
./rsrc -arch 386 -ico icon.ico -o rsrc_386.syso
./rsrc -arch amd64 -ico icon.ico -o rsrc_amd64.syso

# Build the executable.
go build -ldflags "-s -w" -o "No-Brain Jogging"
