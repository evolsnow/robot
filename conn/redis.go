package conn

import (
	"github.com/garyburd/redigo/redis"
	"log"
	"strconv"
)

type Memo struct {
	Time    string `redis:"time"`
	Content string `redis:"content"`
}

type Task struct {
	Id     int    `redis:"id"`
	ChatId int    `redis:"chatId"`
	Owner  string `redis:"owner"`
	Desc   string `redis:"content"`
	When   string `redis:"time"`
}

//All redis CRUD actions

func CreateMasterId(id int) {
	c := Pool.Get()
	defer c.Close()
	c.Do("SET", "evolsnowChatId", id)
}

func ReadMasterId() int {
	c := Pool.Get()
	defer c.Close()
	id, _ := redis.Int(c.Do("GET", "evolsnowChatId"))
	return id
}

func CreateUserChatId(user string, id int) {
	c := Pool.Get()
	defer c.Close()
	key := user + "ChatId"
	c.Do("SET", key, id)
}

func ReadUserChatId(user string) int {
	c := Pool.Get()
	defer c.Close()
	key := user + "ChatId"
	id, _ := redis.Int(c.Do("GET", key))
	return id
}

func CreateMemo(user, time, memo string) {
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

func DeleteMemo(user string, memos []string) {
	c := Pool.Get()
	defer c.Close()
	var deleteMemoLua = `
	local id = redis.call("LINDEX", KEYS[1]..":memos")
	redis.call("LREM", KEYS[1]..":memos", 1, id)
	redis.call("DEL", "memo:"..elem)
	`
	script := redis.NewScript(1, deleteMemoLua)
	for i := range memos {
		index, _ := strconv.Atoi(memos[i])
		script.Do(c, user, index-1)
	}
}

func UpdateTaskId() int {
	c := Pool.Get()
	defer c.Close()
	id, _ := redis.Int(c.Do("INCR", "taskIncrId"))
	return id
}

func CreateTask(ts Task) {
	c := Pool.Get()
	defer c.Close()
	log.Println("save task")
	var setTaskLua = `
	redis.call("RPUSH", "allTasks", KEYS[1])
	redis.call("RPUSH", KEYS[2]..":tasks", KEYS[1])
	redis.call("HMSET", "task:"..KEYS[1], "id", KEYS[1], "owner", KEYS[2], "time", KEYS[3], "content", KEYS[4], "chatID", KEYS[5])
	`
	script := redis.NewScript(5, setTaskLua)
	script.Do(c, ts.Id, ts.Owner, ts.When, ts.Desc, ts.ChatId)
}

func DeleteTask(ts Task) {
	c := Pool.Get()
	defer c.Close()
	log.Println("remove", ts.Id)
	var removeTaskLua = `
	redis.call("LREM", "allTasks", 1, KEYS[2])
	redis.call("LREM", KEYS[1]..":tasks", 1, KEYS[2])
	redis.call("DEL", "task:"..KEYS[2])
	`
	script := redis.NewScript(2, removeTaskLua)
	script.Do(c, ts.Owner, ts.Id)
}

func ReadUserTasks(user string) []Task {
	c := Pool.Get()
	defer c.Close()
	var multiGetTaskLua = `
	local data = redis.call("LRANGE", KEYS[1]..":tasks", "0", "-1")
	local ret = {}
  	for idx=1, #data do
  		ret[idx] = redis.call("HGETALL", "task:"..data[idx])
  	end
  	return ret
   `
	var tasks []Task
	script := redis.NewScript(1, multiGetTaskLua)
	values, err := redis.Values(script.Do(c, user))
	if err != nil {
		log.Println(err)
	}
	for i := range values {
		t := new(Task)
		redis.ScanStruct(values[i].([]interface{}), t)
		tasks = append(tasks, *t)
	}
	return tasks
}

func ReadAllTasks() []Task {
	c := Pool.Get()
	defer c.Close()
	var GetAllTasksLua = `
	local data = redis.call("LRANGE", "allTasks", "0", "-1")
	local ret = {}
  	for idx=1, #data do
  		ret[idx] = redis.call("HGETALL", "task:"..data[idx])
  	end
  	return ret
   `
	var tasks []Task
	script := redis.NewScript(0, GetAllTasksLua)
	values, err := redis.Values(script.Do(c))
	if err != nil {
		log.Println(err)
	}
	for i := range values {
		t := new(Task)
		redis.ScanStruct(values[i].([]interface{}), t)
		tasks = append(tasks, *t)
	}
	return tasks

}

func ReadAllMemos(user string) []Memo {
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
