package conn

import (
	"github.com/garyburd/redigo/redis"
	"log"
)

type Memo struct {
	Time    string `redis:"time"`
	Content string `redis:"content"`
}

type Task struct {
	Id     int
	ChatId int
	Owner  string
	Desc   string
	When   string
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

func HSetTask(ts Task) int {
	c := Pool.Get()
	defer c.Close()
	var setTaskLua = `
	local id = redis.call("INCR", "taskIncrId")
	redis.call("RPUSH", KEYS[1]..":tasks", id)
	redis.call("HMSET", "task:"..id, "time", KEYS[2], "content", KEYS[3], "chatID", KEYS[4])
	return id
	`
	script := redis.NewScript(4, setTaskLua)
	id, _ := redis.Int(script.Do(c, ts.Owner, ts.When, ts.Desc, ts.ChatId))
	return id
}

func RemoveTask(ts Task) {
	c := Pool.Get()
	defer c.Close()
	log.Println(ts.Id)
	var removeTaskLua = `
	redis.call("LREM", KEYS[1]..":tasks", 1, KEYS[2])
	redis.call("DEL", "task:"..KEYS[2])
	`
	script := redis.NewScript(2, removeTaskLua)
	if _, err := script.Do(c, ts.Owner, ts.Id); err != nil {
		log.Println(err)
	}
}

func HGetAllMemos(user string) []Memo {
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
	var memos []Memo
	script := redis.NewScript(1, multiGetMemoLua)
	values, err := redis.Values(script.Do(c, user))
	if err != nil {
		log.Println(err)
	}
	for i := range values {
		m := new(Memo)
		redis.ScanStruct(values[i].([]interface{}), m)
		memos = append(memos, *m)
	}
	return memos
}

//
//var multiGetScript = redis.NewScript(0, multiGetMemoLua)
