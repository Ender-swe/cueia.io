// Command fileservice serves one request from the private SSH subsystem on a workspace VM.
package main

import (
	"github.com/Ender-swe/cueia.io/apps/server/internal/workspace/volume"
	"log"
	"os"
)

func main() {
	s, err := volume.Open(diskMount())
	if err != nil {
		log.Fatal(err)
	}
	defer s.Close()
	if err = s.Serve(os.Stdin, os.Stdout); err != nil {
		log.Fatal(err)
	}
}
