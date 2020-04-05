package index

import (
	"reflect"
	"sync"
	"testing"
)

func TestAddFileInIndex(t *testing.T) {
	expect := ReverseIndex{
		"black": []WordIndex{
			WordIndex{"3.txt", []int{0}},
			WordIndex{"4.txt", []int{0, 4}},
		},
		"cup": []WordIndex{
			WordIndex{"1.txt", []int{0}},
			WordIndex{"2.txt", []int{0}},
		},
		"tea": []WordIndex{
			WordIndex{"1.txt", []int{2}},
			WordIndex{"2.txt", []int{1}},
			WordIndex{"3.txt", []int{1}},
			WordIndex{"4.txt", []int{1, 2, 3}},
		},
	}
	actual := ReverseIndex{
		"cup": []WordIndex{
			WordIndex{"1.txt", []int{0}},
			WordIndex{"2.txt", []int{0}},
		},
		"tea": []WordIndex{
			WordIndex{"1.txt", []int{2}},
			WordIndex{"2.txt", []int{1}},
		},
	}

	mu := &sync.Mutex{}
	wg := &sync.WaitGroup{}

	text := "black tea"

	wg.Add(1)
	actual.addFileInIndex("3.txt", text, mu, wg)

	text = "black tea tea tea black"

	wg.Add(1)
	actual.addFileInIndex("4.txt", text, mu, wg)

	wg.Wait()

	if !reflect.DeepEqual(actual, expect) {
		t.Errorf("\n%v isn't equal to expected\n%v", actual, expect)
	}
}

func TestHandleWords(t *testing.T) {
	in := []string{"hand", "handling", "handle", "to", "i", "I+", "", "his", "hIS", "+mom, ", "+-*/", "-Handling-"}
	expect := []string{"hand", "handl", "handl", "mom", "handl"}
	actual := HandleWords(in)
	if len(expect) != len(actual) {
		t.Errorf("length %v %v isn't equal to expected %v %v", len(actual), actual, len(expect), expect)
	}
	for i := range actual {
		if actual[i] != expect[i] {
			t.Errorf("%v isn't equal to expected %v", actual, expect)
		}
	}
}

func TestHasFileInIndex(t *testing.T) {
	in := []WordIndex{
		WordIndex{"3.txt", []int{0}},
		WordIndex{"1.txt", []int{1}},
	}
	right := "1.txt"
	wrong := "3.json"

	if hasFileInIndex(in, right) == -1 {
		t.Errorf("func didn't find existing file %v", right)
	}
	if hasFileInIndex(in, wrong) != -1 {
		t.Errorf("func found dosen't exist file %v", wrong)
	}
}

func TestSearching(t *testing.T) {
	index := ReverseIndex{
		"black": []WordIndex{
			WordIndex{"3.txt", []int{1}},
			WordIndex{"2.txt", []int{2}},
		},
		"cup": []WordIndex{
			WordIndex{"1.txt", []int{0}},
			WordIndex{"2.txt", []int{0}},
			WordIndex{"3.txt", []int{0}},
		},
		"tea": []WordIndex{
			WordIndex{"1.txt", []int{1}},
			WordIndex{"2.txt", []int{1}},
			WordIndex{"3.txt", []int{2}},
		},
	}
	expect := []string{"3.txt", "2.txt", "1.txt"}

	actual, _ := index.Searching("cup of black tea")

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
