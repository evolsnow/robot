package main

import "testing"

func TestChinese(t *testing.T) {
	heads := []string{"你好 世界", "你好 world", "hello 世界", "1 世界"}
	tails := []string{"hello world", "1 world"}

	for _, word := range heads {
		if !chinese(word) {
			t.Error("should be chinese:", word)
		}
	}
	for _, word := range tails {
		if chinese(word) {
			t.Error("should not be chinese:", word)
		}
	}

}
