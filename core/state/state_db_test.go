package state

import (
	"fmt"
	"sort"
	"testing"
)

type testStruct struct {
	id   int
	name string
}

func Test_search(t *testing.T) {
	revisionId := 3

	testArr := make([]*testStruct, 0)
	testArr = append(testArr, &testStruct{1, "a"})
	testArr = append(testArr, &testStruct{2, "a"})
	testArr = append(testArr, &testStruct{3, "a"})
	idx := sort.Search(len(testArr), func(i int) bool {
		return testArr[i].id >= revisionId
	})

	fmt.Println(idx)
}
