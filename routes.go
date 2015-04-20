package main

import (
	"encoding/json"
	"github.com/fzzy/radix/redis"
	"html/template"
	"net/http"
	"time"
)

func serveHome(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	template.Must(template.New("index.html").Delims("{[{", "}]}").ParseFiles("index.html", "base.html")).Execute(w, r)
}

func serveScripts(w http.ResponseWriter, r *http.Request) {
	RedisClient, err := redis.Dial("tcp", "127.0.0.1:6379")

	if err != nil {
		errHandler(err)
	}

	defer RedisClient.Close()

	w.Header().Set("Content-Type", "application/json")

	scripts, err := RedisClient.Cmd("smembers", "sheriff:scripts").List()

	if err != nil {
		errHandler(err)
	}

	result, _ := json.Marshal(scripts)
	template.Must(template.New("scripts").Parse(string(result[:]))).Execute(w, r)
}

func getScriptInfo(w http.ResponseWriter, r *http.Request) {
	RedisClient, err := redis.Dial("tcp", "127.0.0.1:6379")

	if err != nil {
		errHandler(err)
	}

	defer RedisClient.Close()

	scriptName := r.URL.Path[len("/scripts/"):]

	w.Header().Set("Content-Type", "application/json")

	script, err := RedisClient.Cmd("hgetall", "sheriff:scripts:"+scriptName).Hash()

	if err != nil {
		errHandler(err)
	}

	result, _ := json.Marshal(script)
	template.Must(template.New("scripts").Parse(string(result[:]))).Execute(w, r)

}

func getScriptChart(w http.ResponseWriter, r *http.Request) {
	RedisClient, err := redis.Dial("tcp", "127.0.0.1:6379")

	if err != nil {
		errHandler(err)
	}

	defer RedisClient.Close()

	scriptName := r.URL.Path[len("/scripts/chart/"):]
	queryValues := r.URL.Query()

	if val, ok := queryValues["period"]; ok {
		var d time.Duration = 0
		switch val[0] {
		case "hour":
			d = time.Hour
			break
		case "day":
			d = 24 * time.Hour
			break
		case "week":
			d = 7 * 24 * time.Hour
			break
		case "month":
			d = 30 * 7 * 24 * time.Hour
			break
		}

		t := time.Now().Add(-1 * d).Unix()

		data, err := RedisClient.Cmd("zrangebyscore", "sheriff:"+scriptName, t, "+inf", "withscores").Hash()

		w.Header().Set("Content-Type", "application/json")

		if err != nil {
			errHandler(err)
		}

		result, _ := json.Marshal(data)
		template.Must(template.New("scripts").Parse(string(result[:]))).Execute(w, r)
	} else {
		data, err := RedisClient.Cmd("zrange", "sheriff:"+scriptName, 0, -1, "withscores").Hash()

		w.Header().Set("Content-Type", "application/json")

		if err != nil {
			errHandler(err)
		}

		result, _ := json.Marshal(data)
		template.Must(template.New("scripts").Parse(string(result[:]))).Execute(w, r)

	}

}

func (h *hub) realtimeScripts() {
	RedisClient, err := redis.Dial("tcp", "127.0.0.1:6379")

	if err != nil {
		errHandler(err)
	}

	defer RedisClient.Close()

	for {
		if data, err := RedisClient.Cmd("blpop", "sheriff:realtime", 0).List(); err != nil {
			errHandler(err)
		} else {
			result, _ := json.Marshal(data)
			h.broadcast <- result
		}
	}
}

func serveWebSocket(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", 405)
		return
	}
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		errHandler(err)
		return
	}

	c := &connection{send: make(chan []byte, 256), ws: ws}
	h.register <- c
	go c.writePump()
	go h.realtimeScripts()

	c.readPump()
}
