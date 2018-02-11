package main

import (
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"os"
)

func hashPassword(password string) string {

	hash := sha512.New()
	hash.Write([]byte(password))
	hashString := hex.EncodeToString(hash.Sum(nil))

	return hashString

}

func main() {

	args := os.Args

	if len(args) == 1 {
		fmt.Println("Password is required")
	} else {
		password := os.Args[1]
		fmt.Println(hashPassword(password))

	}

}
