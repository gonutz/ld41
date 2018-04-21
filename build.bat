set GOARCH=386
go build -ldflags "-H=windowsgui -s -w" -o game_no_rsc.exe
blob -folder=rsc -out=rsc.blob
payload -data=rsc.blob -exe=game_no_rsc.exe -output=game.exe
del game_no_rsc.exe
