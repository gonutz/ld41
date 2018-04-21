package main

// file is the function to be called with a file name to create the absolute
// path for assets. Only use paths created with this file.
// In dev mode this will resolve to files in the local resource folder.
// In release mode this will resolve to files in the executable's data blob.
var file func(filename string) string

var cleanUpAssets func() = func() {}
