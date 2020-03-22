package search

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"sort"
	"strings"

	"github.com/kljensen/snowball/english"
	"github.com/polisgo2020/search-tarival/index"
)

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
func Searching(pathToIndex, pathToStopWords string, searchPhrase []string) []string {
	index := readIndex(pathToIndex)
	keywords := convertPharaseToKeywords(searchPhrase, pathToStopWords)

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

	counterUniqueKeywords(results, keywords)
	sortPositions(results)

	for file, result := range results {
		result.maxLengthPhrase = maxLengthSearchPhrase(result.words, keywords)
		results[file] = result
	}

	sliceResults := sortSearchResults(convertMapToSlice(results))

	fmt.Println(sliceResults)
	var searchResult []string

	for _, result := range sliceResults {
		searchResult = append(searchResult, result.file)
	}

	return searchResult
}

func convertPharaseToKeywords(searchPhrase []string, pathToStopWords string) []string {
	var keywords []string

	mapStopWords := index.CreateStopWordsMap(pathToStopWords)

	for _, keyword := range searchPhrase {
		keyword = english.Stem(keyword, false)
		if _, ok := mapStopWords[keyword]; ok {
			continue
		}
		keywords = append(searchPhrase, strings.ToLower(keyword))
	}
	return keywords
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

func sortSearchResults(sliceResults []searchResult) []searchResult {
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
	return sliceResults
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

func readIndex(pathToIndex string) index.ReverseIndex {
	file, err := ioutil.ReadFile(pathToIndex)
	if err != nil {
		log.Fatal(err)
	}

	index := make(index.ReverseIndex)
	if err := json.Unmarshal(file, &index); err != nil {
		log.Fatal(err)
	}
	return index
}
