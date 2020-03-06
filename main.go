package main

import (
	"errors"
	"log"
	"os"
)

// IndexReverse is type for storage reverse index in program
type indexReverse map[string][]string

func main() {
	var path string
	if len(os.Args) < 2 {
		log.Fatal(errors.New("Can't find path to files"))
	} else {
		path = os.Args[1]
	}

	// indexingFolder(path)

	searching(path)
}
