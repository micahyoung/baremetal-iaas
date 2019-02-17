package main

import (
	"fmt"
	"io/ioutil"
	"strings"
)

func main() {
	var err error
	var content []byte

	content, err = ioutil.ReadFile("./test.dir")
	if err != nil {
		panic(err)
	}

	// var entries []string
	// var currentEntry []rune
	// strings.Map(func(r rune) rune {
	// 	if r == 0 || r == 1 || r == 2 {
	// 		if currentEntry != nil {
	// 			entries = append(entries, string(currentEntry))
	// 			currentEntry = nil
	// 		}
	// 	}

	// 	if r >= 32 && r < 127 {
	// 		currentEntry = append(currentEntry, r)
	// 	}
	// 	return -1
	// }, string(content))

	entries := strings.FieldsFunc(string(content), func(r rune) bool {
		if r >= 32 && r < 127 {
			return false
		}

		return true
	})
	fmt.Printf("entries: %#+v\n", entries)
}
