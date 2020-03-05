package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"encoding/json"
)

// Searching is func for search with reverse index
func Searching() {

	file, err := ioutil.ReadFile("output.txt")
	if err != nil {
		log.Fatal(err)
	}

	
	index := make(map[string][]string)
	if err := json.Unmarshal(file, &index); err != nil {
		log.Fatal(err)
	}
	fmt.Println(index)
}
