package main

import (
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

func handler(resp http.ResponseWriter, req *http.Request) {

	req.ParseForm()
	args := req.Form
	password := args["password"][0]
	time.Sleep(5 * time.Second)
	fmt.Fprintf(resp, "%s", hashString(password))

}

func main() {

	http.HandleFunc("/hash", handler)
	http.ListenAndServe(":8080", nil)

}
