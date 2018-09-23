package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"
)

type Configuration struct {
	Port   int
	DBName string
	DBUser string
	DBPass string
}

var Config Configuration

func main() {

	file, _ := os.Open("./config.json")
	decoder := json.NewDecoder(file)
	decoder.Decode(&Config)

	http.HandleFunc("/GetToken", getToken)
	http.HandleFunc("/ShortUrl", shortURL)
	http.HandleFunc("/CreateUser", createUser)
	http.HandleFunc("/", redirectURL)
	err := http.ListenAndServe(":"+strconv.Itoa(Config.Port), nil)
	if err != nil {
		log.Fatal("listen and server", err)
	}
}
