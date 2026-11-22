package main

import "os"

func diskMount() string {
	dir := os.Getenv("CUEIA_DISK_MOUNT")
	if dir == "" {
		dir = "/srv/cueia"
	}
	return dir
}
