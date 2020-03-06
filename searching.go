package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"sort"
	"strings"
)

type searchResult struct {
	name  string
	count uint
}

// Searching is func for search with reverse index
func searching(path string) {

	file, err := ioutil.ReadFile(path + "index.json")
	if err != nil {
		log.Fatal(err)
	}

	index := make(indexReverse)
	if err := json.Unmarshal(file, &index); err != nil {
		log.Fatal(err)
	}

	if len(os.Args) < 3 {
		log.Fatal(errors.New("Can't find keywords"))
	}

	var keywords []string
	for i := 2; i < len(os.Args); i++ {
		keywords = append(keywords, strings.ToLower(os.Args[i]))
	}

	files := map[string]uint{}

	for _, keyword := range keywords {
		if indexBit, ok := index[keyword]; ok {
			for _, file := range indexBit {
				files[file]++
			}
		}
	}

	var results []searchResult

	for name, i := range files {
		result := searchResult{
			name:  name,
			count: i,
		}
		results = append(results, result)
	}
	sort.Slice(results, func(i, j int) bool { return results[i].count > results[j].count })
	fmt.Println(results)
}
