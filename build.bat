set GOARCH=386
go build -ldflags "-H=windowsgui -s -w" -o game.exe
blob -folder=rsc -out=rsc.blob
payload -data=rsc.blob -exe=game.exe -output="No-Brain Jogging.exe"
del game.exe
