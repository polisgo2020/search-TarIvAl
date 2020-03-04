package main

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"unicode"
)

type indexReverse map[string][]string

func main() {
	var path string
	if len(os.Args) < 2 {
		log.Fatal(errors.New("Can't find path to files"))
	} else {
		path = os.Args[1]
	}

	indexingFolder(path)
}

func indexingFolder(path string) {
	files, err := ioutil.ReadDir(path)
	if err != nil {
		log.Fatal(err)
	}
	index := make(indexReverse)

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
	if err := ioutil.WriteFile("output.json", output, 0); err != nil {
		log.Fatal(err)
	}
}
