package main

import (
	"flag"
	"github.com/KarinaMobileSolutions/config"
	"github.com/fzzy/radix/redis"
	"log"
	"strconv"
)

type Conf struct {
	Redis       Database `json:"redis"`
	TemplateDir string   `json:"templateDir"`
	StaticDir   string   `json:"staticDir"`
}

type Database struct {
	Type     string `json:"type"`
	Host     string `json:"host"`
	Port     int64  `json:"port"`
	Username string `json:"username"`
	Password string `json:"password"`
}

var conf Conf

var addr = flag.String("addr", ":8080", "Http serving address")

func errHandler(err error, errType string) {
	if errType == "fatal" {
		log.Fatal(err)
	} else {
		log.Panicln(err)
	}
}

func GetRedisClient() *redis.Client {
	RedisClient, err := redis.Dial(conf.Redis.Type, conf.Redis.Host+":"+strconv.FormatInt(conf.Redis.Port, 10))

	if err != nil {
		errHandler(err, "fatal")
	}

	return RedisClient
}

func main() {
	config.Init(&conf)

	RedisClient := GetRedisClient()

	RedisClient.Cmd("del", "sheriff:realtime")
	RedisClient.Close()

	flag.Parse()
	StartServer()
}
