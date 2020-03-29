package index

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"strings"
	"sync"
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
func CreateStopWordsMap(path string) (map[string]bool, error) {
	fileStopWords, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	stopWords := strings.Fields(string(fileStopWords))

	mapStopWords := map[string]bool{}
	for _, stopWord := range stopWords {
		mapStopWords[strings.ToLower(stopWord)] = true
	}
	return mapStopWords, nil
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

type tokenData struct {
	token    string
	position int
}

// HandleWords - convert words to correct tokens. Trim, ToLower, Stemmer and exception stop words
func HandleWords(words []string, mapStopWords map[string]bool) []string {
	var tokens []string
	for _, word := range words {
		word = strings.TrimFunc(word, func(r rune) bool {
			return !unicode.IsLetter(r) && !unicode.IsNumber(r)
		})
		word = strings.ToLower(word)
		word = english.Stem(word, false)

		if _, ok := mapStopWords[word]; ok || word == "" {
			continue
		}
		tokens = append(tokens, word)
	}
	return tokens
}

type fileData struct {
	name string
	text []byte
}

// IndexingFolder create a file with revrse index
func IndexingFolder(path, pathToStopWords string) (ReverseIndex, error) {
	files, err := ioutil.ReadDir(path)
	if err != nil {
		return nil, err
	}

	mapStopWords, err := CreateStopWordsMap(pathToStopWords)
	if err != nil {
		return nil, err
	}

	index := make(ReverseIndex)

	ch := make(chan fileData)
	errCh := make(chan error)
	for _, file := range files {
		go readFile(path, file.Name(), ch, errCh)
	}

	mu := &sync.Mutex{}
	wg := &sync.WaitGroup{}

	for i := 0; i != len(files); {
		select {
		case file := <-ch:
			wg.Add(1)
			go addFileInIndex(file.name, file.text, mapStopWords, index, mu, wg)
			i++
		case err := <-errCh:
			close(ch)
			close(errCh)
			return nil, err
		}
	}
	wg.Wait()
	return index, nil
}

func hasFileInIndex(sliceIndex []wordIndex, fileName string) (int, bool) {
	for i, indexWord := range sliceIndex {
		if indexWord.File == fileName {
			return i, true
		}
	}
	return -1, false
}

func addFileInIndex(fileName string, fileText []byte, mapStopWords map[string]bool, index ReverseIndex, mu *sync.Mutex, wg *sync.WaitGroup) {
	defer wg.Done()
	words := strings.Fields(string(fileText))
	tokens := HandleWords(words, mapStopWords)
	wordPosition := 0
	mu.Lock()
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
	mu.Unlock()
}

func readFile(path, fileName string, ch chan<- fileData, errCh chan<- error) {
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
