package main

import "testing"

func TestAIReply(t *testing.T) {
	askEn := "test"
	askZh := "测试"
	if tlAI(askEn) == "" || tlAI(askZh) == "" {
		t.Error("tlAI error")
	}
	//if qinAI(askEn) == "" || qinAI(askZh) == "" {
	//	t.Error("qinAI error")
	//}
	if mitAI(askEn) == "" || mitAI(askZh) == "" {
		t.Error("mitAI error")
	}
}
