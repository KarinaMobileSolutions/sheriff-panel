package main

import (
	"flag"
	"github.com/fzzy/radix/redis"
	"log"
)

var addr = flag.String("addr", ":8080", "Http serving address")

func errHandler(err error) {
	log.Panicln(err)
}

func main() {
	RedisClient, err := redis.Dial("tcp", "127.0.0.1:6379")

	if err != nil {
		errHandler(err)
	}

	RedisClient.Cmd("del", "sheriff:realtime")
	RedisClient.Close()

	flag.Parse()
	StartServer()
}
