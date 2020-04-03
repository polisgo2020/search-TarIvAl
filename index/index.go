package index

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"strings"
	"sync"
	"unicode"

	"github.com/kljensen/snowball/english"
	"github.com/zoomio/stopwords"
)

type wordIndex struct {
	File      string
	Positions []int
}

// ReverseIndex is type for storage reverse index in program
type ReverseIndex map[string][]wordIndex

func (index ReverseIndex) addFileInIndex(fileName string, fileText string, mu *sync.Mutex, wg *sync.WaitGroup) {
	defer wg.Done()
	words := strings.Fields(string(fileText))
	tokens := HandleWords(words)
	wordPosition := 0
	mu.Lock()
	for _, word := range tokens {
		if sliceIndex, ok := index[word]; ok {
			if j := hasFileInIndex(sliceIndex, fileName); j != -1 {
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

// ReadIndexJSON - read 'pathToIndex' file and return ReverseIndex
func ReadIndexJSON(pathToIndex string) (ReverseIndex, error) {
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

// HandleWords - convert words to correct tokens. Trim, ToLower, Stemmer and exception stop words
func HandleWords(words []string) []string {
	var tokens []string
	for _, word := range words {
		word = strings.TrimFunc(word, func(r rune) bool {
			return !unicode.IsLetter(r) && !unicode.IsNumber(r)
		})
		word = strings.ToLower(word)
		word = english.Stem(word, false)

		if stopwords.IsStopWord(word) || word == "" {
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
func IndexingFolder(path string) (ReverseIndex, error) {
	files, err := ioutil.ReadDir(path)
	if err != nil {
		return nil, err
	}

	files = deleteDirs(files)

	index := make(ReverseIndex)

	ch := make(chan fileData)
	errCh := make(chan error)
	for _, file := range files {
		go readFile(path, file.Name(), ch, errCh)
	}

	mu := &sync.Mutex{}
	wg := &sync.WaitGroup{}

	defer close(ch)
	defer close(errCh)

	for i := 0; i != len(files); {
		select {
		case file := <-ch:
			wg.Add(1)
			go index.addFileInIndex(file.name, string(file.text), mu, wg)
			i++
		case err := <-errCh:
			return nil, err
		}
	}
	wg.Wait()
	return index, nil
}

func hasFileInIndex(sliceIndex []wordIndex, fileName string) int {
	for i, indexWord := range sliceIndex {
		if indexWord.File == fileName {
			return i
		}
	}
	return -1
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

func deleteDirs(files []os.FileInfo) []os.FileInfo {
	i := 0
	for _, file := range files {
		if !file.IsDir() {
			files[i] = file
			i++
		}
	}
	files = files[:i]
	return files
}
