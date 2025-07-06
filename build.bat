REM Get the necessary build tool.
go install github.com/gonutz/rsrc@latest

REM Build the resource files with the icon so the Go build tool adds it to the executable.
rsrc -arch 386 -ico icon.ico -o rsrc_386.syso
rsrc -arch amd64 -ico icon.ico -o rsrc_amd64.syso

REM Build a 32 bit Windows exe that runs on all Windows from XP up.
set GOARCH=386
go build -ldflags "-H=windowsgui -s -w" -o "No-Brain Jogging.exe"
