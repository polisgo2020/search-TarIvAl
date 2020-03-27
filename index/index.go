package index

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"unicode"

	"github.com/kljensen/snowball/english"
)

type wordIndex struct {
	File      string
	Positions []int
}

// ReverseIndex is type for storage reverse index in program
type ReverseIndex map[string][]wordIndex

// CreateStopWordsMap - create map stopWords
func CreateStopWordsMap(path string) map[string]bool {
	fileStopWords, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatal(err)
	}
	stopWords := strings.Fields(string(fileStopWords))

	mapStopWords := map[string]bool{}
	for _, stopWord := range stopWords {
		mapStopWords[strings.ToLower(stopWord)] = true
	}
	return mapStopWords
}

// ReadIndex - read 'pathToIndex' file and return ReverseIndex
func ReadIndex(pathToIndex string) (ReverseIndex, error) {
	file, err := ioutil.ReadFile(pathToIndex)
	if err != nil {
		return nil, err
	}

	index := make(ReverseIndex)
	if err := json.Unmarshal(file, &index); err != nil {
		return nil, err
	}
	return index, nil
}

type fileData struct {
	name string
	text []byte
}

// IndexingFolder create a file with revrse index
func IndexingFolder(path, pathToStopWords string) ReverseIndex {
	files, err := ioutil.ReadDir(path)
	if err != nil {
		log.Fatal(err)
	}

	mapStopWords := CreateStopWordsMap(pathToStopWords)
	index := make(ReverseIndex)

	ch := make(chan fileData, len(files))
	exitCh := make(chan struct{}, 1)
	for _, file := range files {
		// var fileText []byte
		go readFile(path, file.Name(), ch, err)
		if err != nil {
			log.Fatal(err)
		}
	}
	// exitCh <- struct{}{}

	for i := 0; ; {
		select {
		case <-exitCh:
			return index
		case file := <-ch:
			addFileInIndex(file.name, file.text, mapStopWords, index)
			i++
			if i == len(files) {
				exitCh <- struct{}{}
			}
		}
	}

	// return index
}

func hasFileInIndex(sliceIndex []wordIndex, fileName string) (int, bool) {
	for i, indexWord := range sliceIndex {
		if indexWord.File == fileName {
			return i, true
		}
	}
	return -1, false
}

func addFileInIndex(fileName string, fileText []byte, mapStopWords map[string]bool, index ReverseIndex) {

	words := strings.Fields(string(fileText))

	wordPosition := 0
	for _, word := range words {
		word = strings.TrimFunc(word, func(r rune) bool {
			return !unicode.IsLetter(r) && !unicode.IsNumber(r)
		})
		word = strings.ToLower(word)
		word = english.Stem(word, false)
		if _, ok := mapStopWords[word]; ok || word == "" {
			continue
		}

		if sliceIndex, ok := index[word]; ok {
			if j, ok := hasFileInIndex(sliceIndex, fileName); ok {
				index[word][j].Positions = append(index[word][j].Positions, wordPosition)
				wordPosition++
				continue
			}
		}

		item := wordIndex{
			File:      fileName,
			Positions: []int{wordPosition},
		}
		index[word] = append(index[word], item)
		wordPosition++
	}
}

func readFile(path, fileName string, ch chan fileData, err error) {
	fileText, err := ioutil.ReadFile(path + string(os.PathSeparator) + fileName)
	file := fileData{
		name: fileName,
		text: fileText,
	}
	ch <- file
}

