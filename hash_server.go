package main

import (
	"context"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"net/http"
	"sync"
	"sync/atomic"
	"time"
)

var id_cnt uint64
var hashes sync.Map

func hashString(str string) string {

	hash := sha512.New()
	hash.Write([]byte(str))
	hashStr := hex.EncodeToString(hash.Sum(nil))

	return hashStr
}

func save_hash(id uint64, password string) {

	time.Sleep(5 * time.Second)
	hashed_pass := hashString(password)
	hashes.Store(id, hashed_pass)
	fmt.Println("hashed")
}

func hashHandler(resp http.ResponseWriter, req *http.Request) {

	req.ParseForm()
	args := req.Form
	password := args["password"][0]

	id := atomic.AddUint64(&id_cnt, 1)
	go save_hash(id, password)
	fmt.Fprintf(resp, "%d\n", id)
}

func shutdownHandler(shutdown chan<- bool) func(resp http.ResponseWriter, req *http.Request) {

	return func(resp http.ResponseWriter, req *http.Request) {
		fmt.Println("Shutdown signal recieved: Finishing requests")
		shutdown <- true
	}
}

func main() {

	srv := http.Server{Addr: ":8080"}
	shutdown := make(chan bool, 1)

	go func() {
		http.HandleFunc("/hash", hashHandler)
		http.HandleFunc("/shutdown", shutdownHandler(shutdown))
		srv.ListenAndServe()
	}()

	<-shutdown

	ctx, _ := context.WithTimeout(context.Background(), 6*time.Second)
	srv.Shutdown(ctx)
	fmt.Println("Server shutdown")
}
