package main

import (
	"fmt"
	"log"
	"os"

	"github.com/3ux1n3/agsm/cmd"
)

var version = "dev"

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "version", "--version":
			fmt.Printf("agsm %s\n", version)
			return
		}
	}

	if err := cmd.Run(); err != nil {
		log.Fatal(err)
	}
}
