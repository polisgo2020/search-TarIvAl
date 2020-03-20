package search

import (
	"encoding/json"
	"errors"
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
func Searching(indexFile, pathToStopWords string, sliceKeywords []string) {

	file, err := ioutil.ReadFile(indexFile)
	if err != nil {
		log.Fatal(err)
	}

	index := make(index.ReverseIndex)
	if err := json.Unmarshal(file, &index); err != nil {
		log.Fatal(err)
	}

	if len(sliceKeywords) == 0 {
		log.Fatal(errors.New("Can't find keywords"))
	}

	keywords := createSliceKeywords(sliceKeywords, pathToStopWords)

	results := map[string]searchResult{}

	for _, keyword := range keywords {
		if indexBit, ok := index[keyword]; ok {
			for _, indexFile := range indexBit {

				var result []wordOnFile

				for _, position := range indexFile.Positions {
					result = append(result, wordOnFile{
						word:     keyword,
						position: position,
					})
				}

				if _, ok := results[indexFile.File]; !ok {
					results[indexFile.File] = searchResult{
						count:          len(indexFile.Positions),
						uniqueKeywords: 0,
						words:          result,
					}
				} else {
					Result := results[indexFile.File]
					Result.words = append(Result.words, result...)
					Result.count += len(indexFile.Positions)
					results[indexFile.File] = Result

				}
			}
		}
	}

	counterUniqueKeywords(results, keywords)
	sortSlicePositions(results)

	for file, result := range results {
		result.maxLengthPhrase = maxLengthKeyphrase(result.words, keywords)
		results[file] = result
	}

	output := []searchResult{}

	for file, result := range results {
		output = append(output, searchResult{
			file:            file,
			count:           result.count,
			uniqueKeywords:  result.uniqueKeywords,
			maxLengthPhrase: result.maxLengthPhrase,
		})
	}

	sort.Slice(output, func(i, j int) bool {
		if output[i].maxLengthPhrase > output[j].maxLengthPhrase {
			return true
		}
		if output[i].maxLengthPhrase == output[j].maxLengthPhrase && output[i].uniqueKeywords > output[j].uniqueKeywords {
			return true
		}
		if output[i].maxLengthPhrase == output[j].maxLengthPhrase && output[i].uniqueKeywords == output[j].uniqueKeywords && output[i].count > output[j].count {
			return true
		}
		return false
	})

	fmt.Println(output)
}

func createSliceKeywords(sliceKeywords []string, pathToStopWords string) []string {
	var keywords []string

	mapStopWords := index.CreateStopWordsMap(pathToStopWords)

	for _, keyword := range sliceKeywords {
		keyword = english.Stem(keyword, false)
		if _, ok := mapStopWords[keyword]; ok {
			continue
		}
		keywords = append(keywords, strings.ToLower(keyword))
	}
	return keywords
}

func counterUniqueKeywords(results map[string]searchResult, keywords []string) {
	for i, result := range results {
		for _, keyword := range keywords {
			for _, Word := range result.words {
				if Word.word == keyword {
					result.uniqueKeywords++
					results[i] = result
					break
				}
			}
		}
	}
}

func sortSlicePositions(results map[string]searchResult) {
	for file, result := range results {
		sort.Slice(result.words, func(i, j int) bool { return result.words[i].position < result.words[j].position })
		results[file] = result
	}
}

func maxLengthKeyphrase(words []wordOnFile, keywords []string) int {
	startKeywordPhrasePositon := 0
	length := 0
	maxLength := 1
	prevPosition := words[0].position - 1
	for _, wordData := range words {
		if startKeywordPhrasePositon+length == len(keywords) {
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
