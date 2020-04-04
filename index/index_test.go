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
