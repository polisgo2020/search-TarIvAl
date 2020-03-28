package index

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"sort"
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

	ch := make(chan fileData)
	errCh := make(chan error)
	for _, file := range files {
		go readFile(path, file.Name(), ch, errCh)
	}

	for i := 0; i != len(files); {
		select {
		case file := <-ch:
			addFileInIndex(file.name, file.text, mapStopWords, index)
			i++
		case <-errCh:
			log.Fatal(err)
		}
	}
	close(ch)
	close(errCh)
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

func addFileInIndex(fileName string, fileText []byte, mapStopWords map[string]bool, index ReverseIndex) {

	words := strings.Fields(string(fileText))
	tokens := HandleWords(words, mapStopWords)
	wordPosition := 0
	for _, word := range tokens {

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

	// for i := 0; ; {
	// 	select {
	// 	case <-exitCh:
	// 		return index
	// 	case file := <-ch:
	// 		addFileInIndex(file.name, file.text, mapStopWords, index)
	// 		i++
	// 		if i == len(files) {
	// 			exitCh <- struct{}{}
	// 		}
	// 	}
	// }
}

func readFile(path, fileName string, ch chan fileData, errCh chan error) {
	fileText, err := ioutil.ReadFile(path + string(os.PathSeparator) + fileName)
	if err != nil {
		errCh <- err
		return
	}
	file := fileData{
		name: fileName,
		text: fileText,
	}
	ch <- file
}

type tokenData struct {
	token    string
	position int
}

// HandleWords - convert words to correct tokens. Trim, ToLower, Stemmer and exception stop words
func HandleWords(words []string, mapStopWords map[string]bool) []string {
	tokensCh := make(chan tokenData)

	for i, word := range words {
		go func(position int, word string, ch chan tokenData) {
			word = strings.TrimFunc(word, func(r rune) bool {
				return !unicode.IsLetter(r) && !unicode.IsNumber(r)
			})
			word = strings.ToLower(word)
			word = english.Stem(word, false)
			ch <- tokenData{
				token:    word,
				position: position,
			}
		}(i, word, tokensCh)
	}
	var tokensData []tokenData
	for i := 0; i < len(words); {
		select {
		case token := <-tokensCh:
			tokensData = append(tokensData, token)
			i++

		}
	}
	sort.Slice(tokensData, func(i, j int) bool {
		return tokensData[i].position < tokensData[j].position
	})

	var tokens []string
	for _, token := range tokensData {
		if _, ok := mapStopWords[token.token]; ok || token.token == "" {
			continue
		}
		tokens = append(tokens, token.token)
	}
	return tokens
}
