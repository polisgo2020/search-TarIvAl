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

	pathToStopWords := "stopwords.txt"
	indexName := "index.json"

	switch os.Args[1] {
	case "index":
		if len(os.Args) < 3 {
			log.Fatal(errors.New("Path to folder not found"))
		}

		path = os.Args[2]
		index, err := index.IndexingFolder(path, pathToStopWords)
		if err != nil {
			log.Fatal(err)
		}
		output, err := json.Marshal(index)
		if err != nil {
			log.Fatal(err)
		}
		if err := ioutil.WriteFile(indexName, output, 0); err != nil {
			log.Fatal(err)
		}
	case "search":
		if len(os.Args) < 3 {
			log.Fatal(errors.New("Search phrase not found"))
		}

		mapStopWords, err := index.CreateStopWordsMap(pathToStopWords)
		if err != nil {
			log.Fatal(err)
		}
		keywords := index.HandleWords(os.Args[2:], mapStopWords)
		index, err := index.ReadIndex(indexName)
		if err != nil {
			log.Fatal(err)
		}

		searchResult := search.Searching(index, keywords)
		output := ""
		for i, result := range searchResult {
			output = fmt.Sprintf("%v%v) %v\n", output, i+1, result)
		}

		if err := ioutil.WriteFile("stdout.txt", []byte(output), 0); err != nil {
			log.Fatal(err)
		}

	case "help":
		fmt.Printf("Run:\n1) 'search-tarival index (path to folder for indexing)' for indexing folder\n2) 'search-tarival search (search phrase)' for search in folder with existing index")
	default:
		fmt.Printf("Unknown command.\nRun 'search-tarival help' for usage.")
	}
}
