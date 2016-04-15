package conn

import (
	"github.com/garyburd/redigo/redis"
	"log"
)

// Memo is user's memo
type Memo struct {
	Time    string `redis:"time"`
	Content string `redis:"content"`
}

// Task is user's task
type Task struct {
	Id     int    `redis:"id"`
	ChatId int64  `redis:"chatId"`
	Owner  string `redis:"owner"`
	Desc   string `redis:"content"`
	When   string `redis:"time"`
}

//All redis CRUD actions

// CreateMasterId saves master's id
func CreateMasterId(id int64) {
	c := Pool.Get()
	defer c.Close()
	c.Do("SET", "masterChatId", id)
}

// ReadMasterId read master id in redis
func ReadMasterId() int64 {
	c := Pool.Get()
	defer c.Close()
	id, _ := redis.Int64(c.Do("GET", "masterChatId"))
	return id
}

// CreateUserChatId saves user's chat id
func CreateUserChatId(user string, id int64) {
	c := Pool.Get()
	defer c.Close()
	key := user + "ChatId"
	c.Do("SET", key, id)
}

// ReadUserChatId read user's chat id
func ReadUserChatId(user string) int64 {
	c := Pool.Get()
	defer c.Close()
	key := user + "ChatId"
	id, _ := redis.Int64(c.Do("GET", key))
	return id
}

// CreateMemo saves a memo
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

// DeleteMemo deletes a memo
func DeleteMemo(user string, index int) {
	c := Pool.Get()
	defer c.Close()
	var deleteMemoLua = `
	local id = redis.call("LINDEX", KEYS[1]..":memos", KEYS[2])
	redis.call("LREM", KEYS[1]..":memos", 1, id)
	redis.call("DEL", "memo:"..id)
	`
	script := redis.NewScript(2, deleteMemoLua)
	script.Do(c, user, index)
}

// UpdateTaskId auto increases task id
func UpdateTaskId() int {
	c := Pool.Get()
	defer c.Close()
	id, _ := redis.Int(c.Do("INCR", "taskIncrId"))
	return id
}

// CreateTask saves a task
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

// DeleteTask deletes a task
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

// ReadUserTasks read user's all tasks
func ReadUserTasks(user string) []Task {
	c := Pool.Get()
	defer c.Close()
	var lua = `
	local data = redis.call("LRANGE", KEYS[1]..":tasks", "0", "-1")
	local ret = {}
  	for idx=1, #data do
  		ret[idx] = redis.call("HGETALL", "task:"..data[idx])
  	end
  	return ret
   	`
	var tasks []Task
	script := redis.NewScript(1, lua)
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

// ReadAllTasks load all unfinished task
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

// ReadAllMemo get user's all memos
func ReadAllMemos(user string) []Memo {
	c := Pool.Get()
	defer c.Close()
	var lua = `
	local data = redis.call("LRANGE", KEYS[1]..":memos", "0", "-1")
	local ret = {}
  	for idx=1, #data do
  		ret[idx] = redis.call("HGETALL", "memo:"..data[idx])
  	end
  	return ret
   	`
	var memos []Memo
	script := redis.NewScript(1, lua)
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

// CreateDownloadRecord records user's latest American show download
func CreateDownloadRecord(user, show, se string) {
	c := Pool.Get()
	defer c.Close()
	c.Do("HSET", user+":shows", show, se)
}

// ReadDownloadRecord get latest download record
func ReadDownloadRecord(user, show string) string {
	c := Pool.Get()
	defer c.Close()
	ret, _ := redis.String(c.Do("HGET", user+":shows", show))
	return ret
}
