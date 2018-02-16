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
var shutdown chan bool
var total_request_time uint64

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

func saveHash(id uint64, password string) {

	time.Sleep(5 * time.Second)
	hashed_pass := hashString(password)
	hashes.Store(id, hashed_pass)
	fmt.Println("Password hashed")
}

func updateTotalTime(start_time time.Time) {
	elapsed := time.Since(start_time)
	atomic.AddUint64(&total_request_time, uint64(elapsed))
}

func hashHandler(resp http.ResponseWriter, req *http.Request) {

	start := time.Now()
	defer updateTotalTime(start)
	req.ParseForm()
	args := req.Form

	if len(args["password"]) == 0 {
		fmt.Fprintf(resp, "Error: Missing password in form data\n")
	} else {
		password := args["password"][0]
		id := atomic.AddUint64(&id_cnt, 1)
		go saveHash(id, password)
		fmt.Fprintf(resp, "%d\n", id)
	}
}

func retrieveHandler(resp http.ResponseWriter, req *http.Request) {
	uri := req.RequestURI
	id, err := strconv.ParseUint(path.Base(uri), 10, 64)
	hash, ok := hashes.Load(id)

	if err != nil {
		fmt.Fprintf(resp, "Error: ID not valid format\n")
	} else if ok == true {
		fmt.Fprintf(resp, "%s\n", hash)
	} else {
		fmt.Fprintf(resp, "Error: Hash with ID %d not found\n", id)
	}
}

func shutdownHandler(resp http.ResponseWriter, req *http.Request) {

	fmt.Println("Shutdown signal recieved: Finishing requests")
	shutdown <- true
}

func statsHandler(resp http.ResponseWriter, req *http.Request) {

	total_hashes := atomic.LoadUint64(&id_cnt)
	request_time := atomic.LoadUint64(&total_request_time)

	var avg_request_time float64
	if total_hashes == 0 {
		avg_request_time = 0.0
	} else {
		avg_request_time = (float64(request_time) / float64(total_hashes)) / float64(time.Millisecond)
	}

	stat := Stats{Total: total_hashes, Average: avg_request_time}
	json.NewEncoder(resp).Encode(stat)
}

func main() {

	srv := http.Server{Addr: ":8080"}
	shutdown = make(chan bool, 1)

	go func() {
		http.HandleFunc("/hash/", retrieveHandler)
		http.HandleFunc("/hash", hashHandler)
		http.HandleFunc("/stats", statsHandler)
		http.HandleFunc("/shutdown", shutdownHandler)
		srv.ListenAndServe()
	}()

	<-shutdown

	ctx, _ := context.WithTimeout(context.Background(), 6*time.Second)
	srv.Shutdown(ctx)
	fmt.Println("Server shutdown")
}
