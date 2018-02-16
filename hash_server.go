package main

import (
	"context"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"path"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

var id_cnt uint64
var hashes sync.Map
var total_hash_time uint64

type Stats struct {
	Total   uint64  `json:"total"`
	Average float64 `json:"average"`
}

func hashString(str string) string {

	hash := sha512.New()
	hash.Write([]byte(str))
	hashStr := hex.EncodeToString(hash.Sum(nil))

	return hashStr
}

func save_hash(id uint64, password string) {

	time.Sleep(4 * time.Second)
	hashed_pass := hashString(password)
	hashes.Store(id, hashed_pass)
	fmt.Println("hashed")
}

func hashHandler(resp http.ResponseWriter, req *http.Request) {

	start := time.Now()
	req.ParseForm()
	args := req.Form
	password := args["password"][0]

	id := atomic.AddUint64(&id_cnt, 1)
	go save_hash(id, password)
	elapsed := time.Since(start)
	atomic.AddUint64(&total_hash_time, uint64(elapsed))

	fmt.Fprintf(resp, "%d\n", id)
}

func retrieveHandler(resp http.ResponseWriter, req *http.Request) {
	uri := req.RequestURI
	id, _ := strconv.ParseUint(path.Base(uri), 10, 64)
	hash, _ := hashes.Load(id)
	fmt.Fprintf(resp, "%s\n", hash)
}

func shutdownHandler(shutdown chan<- bool) func(resp http.ResponseWriter, req *http.Request) {

	return func(resp http.ResponseWriter, req *http.Request) {
		fmt.Println("Shutdown signal recieved: Finishing requests")
		shutdown <- true
	}
}

func statsHandler(resp http.ResponseWriter, req *http.Request) {

	total_hashes := atomic.LoadUint64(&id_cnt)
	hash_time := atomic.LoadUint64(&total_hash_time)

	var avg_hash_time float64
	if total_hashes == 0 {
		avg_hash_time = 0.0
	} else {
		avg_hash_time = (float64(hash_time) / float64(total_hashes)) / float64(time.Millisecond)
	}

	stat := Stats{Total: total_hashes, Average: avg_hash_time}
	json.NewEncoder(resp).Encode(stat)
}

func main() {

	srv := http.Server{Addr: ":8080"}
	shutdown := make(chan bool, 1)

	go func() {
		http.HandleFunc("/hash/", retrieveHandler)
		http.HandleFunc("/hash", hashHandler)
		http.HandleFunc("/stats", statsHandler)
		http.HandleFunc("/shutdown", shutdownHandler(shutdown))
		srv.ListenAndServe()
	}()

	<-shutdown

	ctx, _ := context.WithTimeout(context.Background(), 6*time.Second)
	srv.Shutdown(ctx)
	fmt.Println("Server shutdown")
}
