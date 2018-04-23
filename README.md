No-Brain Jogging
================

![Screenshot](https://raw.githubusercontent.com/gonutz/ld41/master/screenshots/screen%2002.png)

This is [my entry](https://ldjam.com/events/ludum-dare/41/no-brain-jogging) for the [Ludum Dare 41](https://ldjam.com/events/ludum-dare/41) game jam. The theme is: "Combine two Incompatible Genres".

My two genres are:

- 2D side-scrolling zombie shooter
- educational math game / brain jogging

In this game you solve math calculations to shoot your rifle and kill some zombies. Kill as many as you can before they eat your brains.

Build Instructions
==================

This game is written completely in [the Go programming language](https://golang.org). Download and install it [from here](https://golang.org/dl) and follow the instructions to set up your `GOPATH` environment variable. If you have problems with that or want to know more about it, [go here](https://github.com/golang/go/wiki/SettingGOPATH).

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

Linux Build
-----------

For the Linux build you need the following C libraries installed: `libgl1-mesa-dev`, `libxrandr-dev`, `libxcursor-dev`, `libxinerama-dev`, `libxi-dev`.

On Debian, Ubuntu, Linux Mint etc. do this:

`sudo apt-get install libgl1-mesa-dev libxrandr-dev libxcursor-dev libxinerama-dev libxi-dev`

Run these commands to build and run the game on Linux:

```
go get -u github.com/gonutz/ld41
cd $GOPATH/src/github.com/gonutz/ld41
./build.sh
./"No-Brain Jogging"
```

![Video](https://raw.githubusercontent.com/gonutz/ld41/master/screenshots/video%2002.gif)