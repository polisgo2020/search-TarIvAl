package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/polisgo2020/search-tarival/index"
	"github.com/polisgo2020/search-tarival/search"
)

func main() {
	var path string

	if len(os.Args) < 2 {
		log.Fatal(errors.New("Command not found"))
	}
	indexName := "index.json"

	switch os.Args[1] {
	case "index":
		if len(os.Args) < 3 {
			log.Fatal(errors.New("Path to folder not found"))
		}

		path = os.Args[2]
		index, err := index.IndexingFolder(path)
		if err != nil {
			log.Fatal(err)
		}
		output, err := json.Marshal(index)
		if err != nil {
			log.Fatal(err)
		}
		if err := ioutil.WriteFile(indexName, output, 0666); err != nil {
			log.Fatal(err)
		}
	case "search":
		if len(os.Args) < 3 {
			log.Fatal(errors.New("Search phrase not found"))
		}

		keywords := index.HandleWords(os.Args[2:])
		if len(keywords) == 0 {
			log.Fatal(errors.New("Search phrase doesn't contain keywords"))
		}
		index, err := index.ReadIndexJSON(indexName)
		if err != nil {
			log.Fatal(err)
		}

		searchResult := search.Searching(index, keywords)
		for i, result := range searchResult {
			fmt.Printf("%v) %v\n", i+1, result)
		}
	case "help":
		fmt.Printf("Run:\n1) 'search-tarival index (path to folder for indexing)' for indexing folder\n2) 'search-tarival search (search phrase)' for search in folder with existing index")
	default:
		fmt.Printf("Unknown command.\nRun 'search-tarival help' for usage.")
	}
}
