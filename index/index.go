package index

import (
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

// IndexingFolder create a file with revrse index
func IndexingFolder(path, pathToStopWords string) ReverseIndex {
	files, err := ioutil.ReadDir(path)
	if err != nil {
		log.Fatal(err)
	}

	mapStopWords := CreateStopWordsMap(pathToStopWords)
	index := make(ReverseIndex)

	for _, file := range files {
		addFileInIndex(file, path, mapStopWords, index)
	}

	return index
}

func hasFileInIndex(sliceIndex []wordIndex, fileName string) (int, bool) {
	for i, indexWord := range sliceIndex {
		if indexWord.File == fileName {
			return i, true
		}
	}
	return -1, false
}

func addFileInIndex(file os.FileInfo, path string, mapStopWords map[string]bool, index ReverseIndex) {
	fileText, err := ioutil.ReadFile(path + string(os.PathSeparator) + file.Name())
	if err != nil {
		log.Fatal(err)
	}

	words := strings.Fields(string(fileText))

	for i, word := range words {
		word = strings.TrimFunc(word, func(r rune) bool {
			return !unicode.IsLetter(r) && !unicode.IsNumber(r)
		})
		word = strings.ToLower(word)
		word = english.Stem(word, false)
		if _, ok := mapStopWords[word]; ok || word == "" {
			continue
		}

		if sliceIndex, ok := index[word]; ok {
			if j, ok := hasFileInIndex(sliceIndex, file.Name()); ok {
				index[word][j].Positions = append(index[word][j].Positions, i)
				continue
			}
		}

		item := wordIndex{
			File:      file.Name(),
			Positions: []int{i},
		}
		index[word] = append(index[word], item)
	}
}
