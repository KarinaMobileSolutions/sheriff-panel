package main

import (
	"flag"
	"log"
)

var addr = flag.String("addr", ":8080", "Http serving address")

func errHandler(err error) {
	log.Panicln(err)
}

func main() {
	flag.Parse()
	StartServer()
}
