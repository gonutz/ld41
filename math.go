package main

import "time"

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func frames(d time.Duration) int {
	return int(60 * d / time.Second)
}
