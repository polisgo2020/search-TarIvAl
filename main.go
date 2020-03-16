package main

import (
	"errors"
	"log"
	"os"
	"github.com/polisgo2020/search-tarival/index"
	"github.com/polisgo2020/search-tarival/search"
)

func main() {
	var path string
	if len(os.Args) < 2 {
		log.Fatal(errors.New("Can't find path to files"))
	} else {
		path = os.Args[1]
	}

	IndexingFolder(path)
	Searching()
}
