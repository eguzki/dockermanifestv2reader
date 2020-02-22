package main

import (
	"log"
	"os"

	"github.com/eguzki/dockermanifestv2reader/pkg/reader"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatalln("image url missing")
	}

	err := reader.Read(os.Args[1])
	if err != nil {
		log.Fatalln(err)
	}
}
