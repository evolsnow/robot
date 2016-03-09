package main

import "unicode"

// if the words contain any chinese character, return true
func chinese(words string) (zh bool) {
	for _, r := range words {
		if unicode.Is(unicode.Scripts["Han"], r) {
			zh = true
			break
		}
	}
	return
}
