package search

import (
	"reflect"
	"testing"

	"github.com/polisgo2020/search-tarival/index"
)

func TestSearching(t *testing.T) {
	Index := index.ReverseIndex{
		"black": []index.WordIndex{
			index.WordIndex{"3.txt", []int{1}},
			index.WordIndex{"2.txt", []int{2}},
		},
		"cup": []index.WordIndex{
			index.WordIndex{"1.txt", []int{0}},
			index.WordIndex{"2.txt", []int{0}},
			index.WordIndex{"3.txt", []int{0}},
		},
		"tea": []index.WordIndex{
			index.WordIndex{"1.txt", []int{1}},
			index.WordIndex{"2.txt", []int{1}},
			index.WordIndex{"3.txt", []int{2}},
		},
	}
	expect := []string{"3.txt", "2.txt", "1.txt"}

	actual, _ := Searching(Index, "cup of black tea")

	if !reflect.DeepEqual(actual, expect) {
		t.Errorf("\n%v isn't equal to expected\n%v", actual, expect)
	}
}

func TestCounterUniqueKeywords(t *testing.T) {
	actual := map[string]searchResult{
		"1.txt": searchResult{
			count:          2,
			uniqueKeywords: 0,
			words: []wordOnFile{
				wordOnFile{
					word:     "cup",
					position: 0,
				},
				wordOnFile{
					word:     "tea",
					position: 1,
				},
			},
		},
		"2.txt": searchResult{
			count:          3,
			uniqueKeywords: 0,
			words: []wordOnFile{
				wordOnFile{
					word:     "cup",
					position: 0,
				},
				wordOnFile{
					word:     "tea",
					position: 1,
				},
				wordOnFile{
					word:     "black",
					position: 2,
				},
			},
		},
		"3.txt": searchResult{
			count:          3,
			uniqueKeywords: 0,
			words: []wordOnFile{
				wordOnFile{
					word:     "cup",
					position: 0,
				},
				wordOnFile{
					word:     "black",
					position: 1,
				},
				wordOnFile{
					word:     "tea",
					position: 2,
				},
			},
		},
	}
	expect := map[string]searchResult{
		"1.txt": searchResult{
			count:          2,
			uniqueKeywords: 2,
			words: []wordOnFile{
				wordOnFile{
					word:     "cup",
					position: 0,
				},
				wordOnFile{
					word:     "tea",
					position: 1,
				},
			},
		},
		"2.txt": searchResult{
			count:          3,
			uniqueKeywords: 3,
			words: []wordOnFile{
				wordOnFile{
					word:     "cup",
					position: 0,
				},
				wordOnFile{
					word:     "tea",
					position: 1,
				},
				wordOnFile{
					word:     "black",
					position: 2,
				},
			},
		},
		"3.txt": searchResult{
			count:          3,
			uniqueKeywords: 3,
			words: []wordOnFile{
				wordOnFile{
					word:     "cup",
					position: 0,
				},
				wordOnFile{
					word:     "black",
					position: 1,
				},
				wordOnFile{
					word:     "tea",
					position: 2,
				},
			},
		},
	}

	in := []string{"cup", "black", "tea"}

	counterUniqueKeywords(actual, in)

	if !reflect.DeepEqual(actual, expect) {
		t.Errorf("\n%v isn't equal to expected\n%v", actual, expect)
	}
}
func TestMaxLengthSearchPhrase(t *testing.T) {
	actual := map[string]searchResult{
		"1.txt": searchResult{
			count:          2,
			uniqueKeywords: 2,
			words: []wordOnFile{
				wordOnFile{
					word:     "cup",
					position: 0,
				},
				wordOnFile{
					word:     "tea",
					position: 1,
				},
			},
		},
		"2.txt": searchResult{
			count:          3,
			uniqueKeywords: 3,
			words: []wordOnFile{
				wordOnFile{
					word:     "cup",
					position: 0,
				},
				wordOnFile{
					word:     "tea",
					position: 1,
				},
				wordOnFile{
					word:     "black",
					position: 2,
				},
			},
		},
		"3.txt": searchResult{
			count:          3,
			uniqueKeywords: 3,
			words: []wordOnFile{
				wordOnFile{
					word:     "cup",
					position: 0,
				},
				wordOnFile{
					word:     "black",
					position: 1,
				},
				wordOnFile{
					word:     "tea",
					position: 2,
				},
			},
		},
	}
	expect := map[string]searchResult{
		"1.txt": searchResult{
			count:           2,
			uniqueKeywords:  2,
			maxLengthPhrase: 1,
			words: []wordOnFile{
				wordOnFile{
					word:     "cup",
					position: 0,
				},
				wordOnFile{
					word:     "tea",
					position: 1,
				},
			},
		},
		"2.txt": searchResult{
			count:           3,
			uniqueKeywords:  3,
			maxLengthPhrase: 1,
			words: []wordOnFile{
				wordOnFile{
					word:     "cup",
					position: 0,
				},
				wordOnFile{
					word:     "tea",
					position: 1,
				},
				wordOnFile{
					word:     "black",
					position: 2,
				},
			},
		},
		"3.txt": searchResult{
			count:           3,
			uniqueKeywords:  3,
			maxLengthPhrase: 3,
			words: []wordOnFile{
				wordOnFile{
					word:     "cup",
					position: 0,
				},
				wordOnFile{
					word:     "black",
					position: 1,
				},
				wordOnFile{
					word:     "tea",
					position: 2,
				},
			},
		},
	}

	in := []string{"cup", "black", "tea"}

	for file, result := range actual {
		result.maxLengthPhrase = maxLengthSearchPhrase(result.words, in)
		actual[file] = result
	}

	if !reflect.DeepEqual(actual, expect) {
		t.Errorf("\n%v isn't equal to expected\n%v", actual, expect)
	}
}
