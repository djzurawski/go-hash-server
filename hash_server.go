package main

import (
	"context"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"net/http"
	"time"
)

func hashString(str string) string {

	hash := sha512.New()
	hash.Write([]byte(str))
	hashStr := hex.EncodeToString(hash.Sum(nil))

	return hashStr
}

func hashHandler(resp http.ResponseWriter, req *http.Request) {

	req.ParseForm()
	args := req.Form
	password := args["password"][0]
	time.Sleep(5 * time.Second)
	fmt.Fprintf(resp, "%s", hashString(password))
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

	ctx, _ := context.WithTimeout(context.Background(), 6 * time.Second)
	srv.Shutdown(ctx)
	fmt.Println("Server shutdown")
}
