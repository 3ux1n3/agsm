package main

import (
	"log"

	"github.com/3ux1n3/agsm/cmd"
)

func main() {
	if err := cmd.Run(); err != nil {
		log.Fatal(err)
	}
}
