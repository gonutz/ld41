REM build the tools that are necessary to build the final game
go build ./vendor/github.com/gonutz/blob/cmd/blob
go build ./vendor/github.com/gonutz/payload/cmd/payload
go build ./vendor/github.com/gonutz/rsrc

REM build the resource files with the icon so the Go build tool adds it to the .exe
rsrc -arch 386 -ico icon.ico -o rsrc_386.syso
rsrc -arch amd64 -ico icon.ico -o rsrc_amd64.syso

REM build a 32 bit Windows exe that runs on all Windows from XP up
set GOARCH=386
go build -ldflags "-H=windowsgui -s -w" -o game.exe
blob -folder=rsc -out=rsc.blob
payload -data=rsc.blob -exe=game.exe -output="No-Brain Jogging.exe"

REM delete the temporary helper files
del rsc.blob
del game.exe
del blob.exe
del payload.exe
del rsrc.exe
