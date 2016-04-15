package main

import (
	"encoding/json"
	"io/ioutil"
	"os"
)

// Config for initialized robot and server
type Config struct {
	Server  string `json:"server"`
	Port    int    `json:"port"`
	Cert    string `json:"cert"`
	CertKey string `json:"cert_key"`

	WebHookUrl    string `json:"webhook_url"`
	RedisAddress  string `json:"redis_address"`
	RedisPort     int    `json:"redis_port"`
	RedisDB       int    `json:"redis_db"`
	RedisPassword string `json:"redis_password"`
	RobotName     string `json:"robot_name"`
	RobotToken    string `json:"robot_token"`
}

// ParseConfig parses config from the given file path
func ParseConfig(path string) (config *Config, err error) {
	file, err := os.Open(path)
	if err != nil {
		return
	}
	defer file.Close()
	data, err := ioutil.ReadAll(file)
	if err != nil {
		return
	}
	config = &Config{}
	if err = json.Unmarshal(data, config); err != nil {
		return nil, err
	}
	return
}
