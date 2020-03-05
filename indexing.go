package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"strings"
	"unicode"
)

// IndexingFolder create a file with revrse index
func IndexingFolder(path string) {
	files, err := ioutil.ReadDir(path)
	if err != nil {
		log.Fatal(err)
	}
	index := make(IndexReverse)

	for _, f := range files {
		file, err := ioutil.ReadFile(path + "\\" + f.Name())
		if err != nil {
			log.Fatal(err)
		}

		words := strings.Fields(string(file))

		for _, word := range words {
			word = strings.TrimFunc(string(word), func(r rune) bool {
				return !unicode.IsLetter(r) && !unicode.IsNumber(r)
			})
			word = strings.ToLower(word)
			index[word] = append(index[word], f.Name())
		}
	}

	output, err := json.Marshal(index)
	if err != nil {
		log.Fatal(err)
	}
	if err := ioutil.WriteFile("index.json", output, 0); err != nil {
		log.Fatal(err)
	}
}
