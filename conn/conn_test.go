package conn

import (
	"testing"
)

var user = "TestUser"

func TestRedisConn(t *testing.T) {
	Pool = NewPool("127.0.0.1:6379", "", 10)
	if Pool == nil {
		t.Error("create new pool error")
	}
	if !Ping("127.0.0.1:6379", "") {
		t.Fatal("connect to redis server failed")
	}
}

func TestCRMasterId(t *testing.T) {
	CreateMasterId(123)
	if ReadMasterId() != 123 {
		t.Error("CR master id error")
	}
}

func TestCRUserChatId(t *testing.T) {
	CreateUserChatId(user, 1234)
	if ReadUserChatId(user) != 1234 {
		t.Error("CR user chat id error")
	}
}
