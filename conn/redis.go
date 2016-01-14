package conn

import (
	"github.com/garyburd/redigo/redis"
)

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
