package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

type Result struct {
	Index int
	Value int
}

var seen []Result

func main() {
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/calc", calculateHandler)

	log.Println("server started on :80")
	log.Fatal(http.ListenAndServe(":80", nil))
}

func indexHandler(w http.ResponseWriter, _ *http.Request) {
	render(w)
}

func calculateHandler(w http.ResponseWriter, r *http.Request) {
	srvHost := os.Getenv("SERVER_HOST")
	srvPort := os.Getenv("SERVER_PORT")

	if srvHost == "" || srvPort == "" {
		srvHost = "localhost"
		srvPort = "3001"
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	index, err := strconv.Atoi(r.FormValue("index"))

	req := struct {
		Num int `json:"num"`
	}{
		Num: index,
	}
	bb, err := json.Marshal(req)

	resp, err := http.Post(fmt.Sprintf("http://%s:%s/add", srvHost, srvPort), "application/json", bytes.NewBuffer(bb))
	if err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	defer resp.Body.Close()

	_, err = io.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	time.Sleep(time.Second * 3)
	respGet, err := http.Get(fmt.Sprintf("http://%s:%s/get/%v", srvHost, srvPort, index))

	if err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	defer respGet.Body.Close()
	res1 := struct {
		Num int
		Fib int
	}{}

	bb, err = io.ReadAll(respGet.Body)
	if err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	if err := json.Unmarshal(bb, &res1); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	//
	seen = append(seen, Result{
		Index: res1.Num,
		Value: res1.Fib,
	})

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func fib(n int) int {
	if n <= 1 {
		return n
	}

	a, b := 0, 1
	for i := 2; i <= n; i++ {
		a, b = b, a+b
	}

	return b
}

func render(w http.ResponseWriter) {
	tmpl := template.Must(template.ParseFiles("templates/main.html"))

	if err := tmpl.Execute(w, seen); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
