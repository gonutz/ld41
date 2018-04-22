No-Brain Jogging
================

This is my entry for the [Ludum Dare 41](https://ldjam.com/events/ludum-dare/41) game jam. The theme is: "Combine two Incompatible Genres".

My two genres are:

- 2D side-scrolling zombie shooter
- educational math game / brain jogging

In this game you solve math calculations to shoot your rifle and kill some zombies. Kill as many as you can before they eat your brains.

Build Instructions
==================

This game is written completely in [the Go programming language](https://golang.org). Download and install it [from here]().

You also need [git](https://git-scm.com/downloads) installed.

Windows Build
-------------

Run these commands to build and run the game on Windows:

```
go get -u github.com/gonutz/ld41
cd %GOPATH%\src\github.com\gonutz\ld41
build.bat
"No-Brain Jogging.exe"
```

This will create and run a statically linked executable containing all the resource data for the game. You can copy the file `No-Brain Jogging.exe` to any Windows machine from XP up and it will run there.