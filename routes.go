package main

import (
	"encoding/json"
	auth "github.com/abbot/go-http-auth"
	"html/template"
	"net/http"
	"time"
)

type ScriptInfo struct {
	Args        string            `json:"args"`
	Cmd         string            `json:"cmd"`
	Description string            `json:"description"`
	Directory   string            `json:"directory"`
	Format      string            `json:"format"`
	Status      map[string]string `json:"status"`
	StatusSort  string            `json:"status_sort"`
}

func serveHome(w http.ResponseWriter, r *auth.AuthenticatedRequest) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	template.Must(template.New("index.html").Delims("{[{", "}]}").ParseFiles(conf.TemplateDir+"/index.html", conf.TemplateDir+"/base.html")).Execute(w, r)
}

func serveScripts(w http.ResponseWriter, r *auth.AuthenticatedRequest) {
	RedisClient := GetRedisClient()

	defer RedisClient.Close()

	w.Header().Set("Content-Type", "application/json")

	scripts, err := RedisClient.Cmd("smembers", "sheriff:scripts").List()

	if err != nil {
		errHandler(err, "panic")
	}

	result, _ := json.Marshal(scripts)
	template.Must(template.New("scripts").Parse(string(result[:]))).Execute(w, r)
}

func getScriptInfo(w http.ResponseWriter, r *auth.AuthenticatedRequest) {
	RedisClient := GetRedisClient()

	defer RedisClient.Close()

	scriptName := r.URL.Path[len("/scripts/"):]

	w.Header().Set("Content-Type", "application/json")

	script, err := RedisClient.Cmd("hgetall", "sheriff:scripts:"+scriptName).Hash()

	if err != nil {
		errHandler(err, "panic")
	}

	scriptStatus, err := RedisClient.Cmd("hgetall", "sheriff:scripts:"+scriptName+":status").Hash()

	if err != nil {
		errHandler(err, "panic")
	}

	info := ScriptInfo{
		Cmd:         script["cmd"],
		Args:        script["args"],
		Directory:   script["directory"],
		Description: script["description"],
		Format:      script["format"],
		Status:      scriptStatus,
		StatusSort:  script["status_sort"],
	}

	result, _ := json.Marshal(info)

	template.Must(template.New("scripts").Parse(string(result[:]))).Execute(w, r)

}

func getScriptChart(w http.ResponseWriter, r *auth.AuthenticatedRequest) {
	RedisClient := GetRedisClient()

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
			errHandler(err, "panic")
		}

		result, _ := json.Marshal(data)
		template.Must(template.New("scripts").Parse(string(result[:]))).Execute(w, r)
	} else {
		data, err := RedisClient.Cmd("zrange", "sheriff:"+scriptName, 0, -1, "withscores").Hash()

		w.Header().Set("Content-Type", "application/json")

		if err != nil {
			errHandler(err, "panic")
		}

		result, _ := json.Marshal(data)
		template.Must(template.New("scripts").Parse(string(result[:]))).Execute(w, r)

	}

}

func (h *hub) realtimeScripts() {
	RedisClient := GetRedisClient()

	defer RedisClient.Close()

	for {
		if data, err := RedisClient.Cmd("blpop", "sheriff:realtime", 0).List(); err != nil {
			errHandler(err, "panic")
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
		errHandler(err, "fatal")
		return
	}

	c := &connection{send: make(chan []byte, 256), ws: ws}
	h.register <- c
	go c.writePump()
	go h.realtimeScripts()

	c.readPump()
}
