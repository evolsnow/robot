package conn

import (
	"github.com/garyburd/redigo/redis"
)

type Memos struct {
	Time    string
	Content string
}

//All redis actions

func SetMasterId(id int) {
	c := Pool.Get()
	defer c.Close()
	c.Do("SET", "evolsnowChatId", id)
}

func GetMasterId() int {
	c := Pool.Get()
	defer c.Close()
	id, _ := redis.Int(c.Do("GET", "evolsnowChatId"))
	return id
}

func SetUserChatId(user string, id int) {
	c := Pool.Get()
	defer c.Close()
	key := user + "ChatId"
	c.Do("SET", key, id)
}

func GetUserChatId(user string) int {
	c := Pool.Get()
	defer c.Close()
	key := user + "ChatId"
	id, _ := redis.Int(c.Do("GET", key))
	return id
}

func HSetMemo(user, time, memo string) {
	c := Pool.Get()
	defer c.Close()
	var setMemoLua = `
	local id = redis.call("INCR", "memoIncrId")
	redis.call("RPUSH", KEYS[1]..":memos", id)
	redis.call("HMSET", "memo:"..id, "time", KEYS[2], "content", KEYS[3])
	`
	script := redis.NewScript(3, setMemoLua)
	script.Do(c, user, time, memo)
}

func HGetAllMemos(user string) *[]Memos {
	c := Pool.Get()
	defer c.Close()
	var multiGetMemoLua = `
	local data = redis.call("LRANGE", KEYS[1]..":memos", "0", "-1")
	local ret = {}
  	for idx=1, #data do
  		ret[idx] = redis.call("HGETALL", "memo:"..data[idx])
  	end
  	return ret
   `
	memos := []Memos{}
	script := redis.NewScript(1, multiGetMemoLua)
	values, _ := redis.Values(script.Do(c, user))
	redis.ScanSlice(values, &memos)
	return &memos
}

//
//var multiGetScript = redis.NewScript(0, multiGetMemoLua)
