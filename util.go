package main

import "unicode"

func chinese(words string) (zh bool) {
	for _, r := range words {
		if unicode.Is(unicode.Scripts["Han"], r) {
			zh = true
			break
		}
	}
	return
}
