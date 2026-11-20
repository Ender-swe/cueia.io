// Command fileservice serves one request from the private SSH subsystem on a workspace VM.
package main

import (
	"github.com/Ender-swe/cueia.io/apps/server/internal/workspace/volume"
	"log"
	"os"
)

func main() {
	dir := os.Getenv("CUEIA_DISK_MOUNT")
	if dir == "" {
		dir = "/srv/cueia"
	}
	s, err := volume.Open(dir)
	if err != nil {
		log.Fatal(err)
	}
	defer s.Close()
	if err = s.Serve(os.Stdin, os.Stdout); err != nil {
		log.Fatal(err)
	}
}
