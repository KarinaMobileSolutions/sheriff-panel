package main

import (
	"github.com/gorilla/mux"
	"log"
	"net/http"
)

func StartServer() {
	r := mux.NewRouter()
	r.HandleFunc("/", serveHome)
	r.PathPrefix("/static/").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, r.URL.Path[1:])
	})
	r.HandleFunc("/scripts", serveScripts)
	r.HandleFunc("/scripts/{script}", getScriptInfo)
	r.HandleFunc("/scripts/chart/{script}", getScriptChart)
	http.Handle("/", r)
	err := http.ListenAndServe(*addr, nil)
	if err != nil {
		log.Fatal(err)
	}
}
