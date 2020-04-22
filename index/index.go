package index

import (
	"database/sql"
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"unicode"

	"github.com/kljensen/snowball/english"
	_ "github.com/lib/pq"
	"github.com/polisgo2020/search-tarival/model"
	"github.com/rs/zerolog/log"
	"github.com/zoomio/stopwords"
)

// indexing

// WordIndex - part index for positions word in one file
type WordIndex struct {
	File      string
	Positions []int
}

// ReverseIndex is type for storage reverse index in program
type ReverseIndex map[string][]WordIndex

type fileData struct {
	name string
	text []byte
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
		word = strings.Replace(word, "'", "", 1)
		tokens = append(tokens, word)
	}
	return tokens
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

func (index ReverseIndex) addFileInIndex(fileName string, fileText string, mu *sync.Mutex, wg *sync.WaitGroup) {
	defer wg.Done()
	words := strings.Fields(string(fileText))
	tokens := HandleWords(words)
	wordPosition := 0
	mu.Lock()
	for _, word := range tokens {
		if sliceIndex, ok := index[word]; ok {
			if j := HasFileInIndex(sliceIndex, fileName); j != -1 {
				index[word][j].Positions = append(index[word][j].Positions, wordPosition)
				wordPosition++
				continue
			}
		}
		item := WordIndex{
			File:      fileName,
			Positions: []int{wordPosition},
		}
		index[word] = append(index[word], item)
		wordPosition++
	}
	mu.Unlock()
}

func readDirInChan(path string, ch chan<- fileData, errCh chan<- error) (int, error) {
	files, err := ioutil.ReadDir(path)
	if err != nil {
		return 0, err
	}

	files = deleteDirs(files)

	for _, file := range files {
		go readFile(path, file.Name(), ch, errCh)
	}
	return len(files), nil
}

// IndexingFolder create a file with reverse index
func IndexingFolder(path string) (ReverseIndex, error) {
	ch := make(chan fileData)
	errCh := make(chan error)
	defer close(ch)
	defer close(errCh)

	countFiles, err := readDirInChan(path, ch, errCh)
	if err != nil {
		return nil, err
	}

	index := make(ReverseIndex)
	mu := &sync.Mutex{}
	wg := &sync.WaitGroup{}

	for i := 0; i != countFiles; {
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

func addFileInDB(db *sql.DB, fileName string, fileText string) error {
	var fID int

	fID, ok, err := model.CheckAndInsert(db, "files", "name_file", "f_id", fileName)
	if err != nil {
		return err
	}
	if ok {
		if err = model.Delete(db, "positions", "f_id", strconv.Itoa(fID)); err != nil {
			return err
		}
	}

	tokens := HandleWords(strings.Fields(string(fileText)))
	wordPosition := 0
	words := make(map[string]int)

	var buffer [][]string
	for _, token := range tokens {
		if _, ok := words[token]; !ok {
			words[token], _, err = model.CheckAndInsert(db, "words", "word", "w_id", token)
			if err != nil {
				return err
			}
		}
		buffer = append(buffer, []string{strconv.Itoa(words[token]), strconv.Itoa(fID), strconv.Itoa(wordPosition)})
		wordPosition++
	}
	if err := model.Insert(db, "positions", []string{"w_id", "f_id", "position"}, buffer); err != nil {
		return err
	}
	log.Info().Str("File", fileName).Msg("File is indexed")
	return nil
}

func readDir(path string) ([]fileData, error) {
	files, err := ioutil.ReadDir(path)
	if err != nil {
		return nil, err
	}

	files = deleteDirs(files)

	ch := make(chan fileData)
	defer close(ch)
	errCh := make(chan error)
	defer close(errCh)
	for _, file := range files {
		go readFile(path, file.Name(), ch, errCh)
	}
	var filesData []fileData
	for i := 0; i != len(files); {
		select {
		case err = <-errCh:
			return nil, err
		case file := <-ch:
			filesData = append(filesData, file)
			i++
		}
	}
	return filesData, nil
}

// IndexingFolderDB save reverse index in db
func IndexingFolderDB(db *sql.DB, path string) error {
	files, err := readDir(path)
	if err != nil {
		return err
	}

	for _, file := range files {
		if err := addFileInDB(db, file.name, string(file.text)); err != nil {
			return err
		}
	}
	return nil
}

// HasFileInIndex find in slice WordIndexs file and returning index for slice item with file
func HasFileInIndex(sliceIndex []WordIndex, fileName string) int {
	for i, indexWord := range sliceIndex {
		if indexWord.File == fileName {
			return i
		}
	}
	return -1
}

// searching
type searchResult struct {
	file            string
	count           int
	uniqueKeywords  int
	maxLengthPhrase int
	words           []wordOnFile
}

type wordOnFile struct {
	word     string
	position int
}

// Searching is func for search with reverse index
func (index ReverseIndex) Searching(searchPhrase string) ([]string, error) {
	keywords := strings.Fields(searchPhrase)
	keywords = HandleWords(keywords)

	if len(keywords) == 0 {
		return nil, errors.New("Search phrase doesn't contain right keywords")
	}

	results := map[string]searchResult{}

	for _, keyword := range keywords {
		if keywordIndex, ok := index[keyword]; ok {
			for _, indexFile := range keywordIndex {
				var words []wordOnFile

				for _, position := range indexFile.Positions {
					words = append(words, wordOnFile{
						word:     keyword,
						position: position,
					})
				}

				if _, ok := results[indexFile.File]; !ok {
					results[indexFile.File] = searchResult{
						count:          len(indexFile.Positions),
						uniqueKeywords: 0,
						words:          words,
					}
				} else {
					result := results[indexFile.File]
					result.words = append(result.words, words...)
					result.count += len(indexFile.Positions)
					results[indexFile.File] = result
				}
			}
		}
	}

	searchResult := handleResults(results, keywords)

	return searchResult, nil
}

// SearchingDB is func for search with reverse index
func SearchingDB(db *sql.DB, searchPhrase string) ([]string, error) {
	keywords := strings.Fields(searchPhrase)
	keywords = HandleWords(keywords)

	if len(keywords) == 0 {
		return nil, errors.New("Search phrase doesn't contain right keywords")
	}

	results := map[string]searchResult{}
	files := make(map[int]string)

	for _, keyword := range keywords {
		var wID int
		switch err := db.QueryRow(`SELECT w_id FROM words WHERE word=$1`, keyword).Scan(&wID); err {
		case sql.ErrNoRows:
			continue
		case nil:
		default:
			return nil, err
		}

		rows, err := db.Query("SELECT f_id, position FROM positions WHERE w_id=$1", wID)
		if err != nil {
			return nil, err
		}
		defer rows.Close()
		for rows.Next() {
			var fID, position int
			err = rows.Scan(&fID, &position)
			if err != nil {
				return nil, err
			}

			if file, ok := files[fID]; !ok {
				err := db.QueryRow(`SELECT name_file FROM files WHERE f_id=$1`, fID).Scan(&file)
				if err != nil {
					return nil, err
				}
				files[fID] = file
			}

			word := wordOnFile{
				word:     keyword,
				position: position,
			}

			if _, ok := results[files[fID]]; !ok {
				results[files[fID]] = searchResult{
					count:          1,
					uniqueKeywords: 0,
					words:          []wordOnFile{word},
				}
			} else {
				result := results[files[fID]]
				result.count++
				result.words = append(result.words, word)
				results[files[fID]] = result
			}
		}
	}

	searchResult := handleResults(results, keywords)

	return searchResult, nil
}

func handleResults(results map[string]searchResult, keywords []string) []string {
	counterUniqueKeywords(results, keywords)
	sortPositions(results)

	for file, result := range results {
		result.maxLengthPhrase = maxLengthSearchPhrase(result.words, keywords)
		results[file] = result
	}

	sliceResults := convertMapToSlice(results)

	sortSearchResults(sliceResults)

	var searchResult []string

	for _, result := range sliceResults {
		searchResult = append(searchResult, result.file)
	}

	return searchResult
}

func counterUniqueKeywords(results map[string]searchResult, keywords []string) {
	for i, result := range results {
		for _, keyword := range keywords {
			for _, Word := range result.words {
				if Word.word == keyword {
					result.uniqueKeywords++
					break
				}
			}
		}
		results[i] = result
	}
}

func sortPositions(results map[string]searchResult) {
	for file, result := range results {
		sort.Slice(result.words, func(i, j int) bool { return result.words[i].position < result.words[j].position })
		results[file] = result
	}
}

func maxLengthSearchPhrase(words []wordOnFile, keywords []string) int {
	startKeywordPhrasePositon := 0
	length := 0
	maxLength := 1
	prevPosition := words[0].position - 1
	for _, wordData := range words {
		if startKeywordPhrasePositon+length >= len(keywords) {
			return maxLength
		}
		if wordData.word == keywords[startKeywordPhrasePositon+length] && wordData.position-1 == prevPosition {
			length++
			if length > maxLength {
				maxLength = length
			}
			if length == len(keywords) {
				return maxLength
			}
		} else {
			length = 0
			for i, keyword := range keywords {
				if wordData.word == keyword {
					length = 1
					startKeywordPhrasePositon = i
					break
				}
			}
		}
		prevPosition = wordData.position
	}
	return maxLength
}

func convertMapToSlice(mapResults map[string]searchResult) []searchResult {
	sliceResults := []searchResult{}
	for file, result := range mapResults {
		sliceResults = append(sliceResults, searchResult{
			file:            file,
			count:           result.count,
			uniqueKeywords:  result.uniqueKeywords,
			maxLengthPhrase: result.maxLengthPhrase,
		})
	}
	return sliceResults
}

func sortSearchResults(sliceResults []searchResult) {
	sort.Slice(sliceResults, func(i, j int) bool {
		if sliceResults[i].maxLengthPhrase > sliceResults[j].maxLengthPhrase {
			return true
		}
		if sliceResults[i].maxLengthPhrase == sliceResults[j].maxLengthPhrase && sliceResults[i].uniqueKeywords > sliceResults[j].uniqueKeywords {
			return true
		}
		if sliceResults[i].maxLengthPhrase == sliceResults[j].maxLengthPhrase && sliceResults[i].uniqueKeywords == sliceResults[j].uniqueKeywords && sliceResults[i].count > sliceResults[j].count {
			return true
		}
		return false
	})
}
