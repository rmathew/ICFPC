package main

import (
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
)

func main() {
	serverURL := os.Args[1]
	playerKey := os.Args[2]

	log.Printf("ServerUrl: %s; PlayerKey: %s", serverURL, playerKey)

	res, err := http.Post(serverURL, "text/plain", strings.NewReader(playerKey))
	if err != nil {
		log.Printf("Unexpected server response:\n%v", err)
		os.Exit(1)
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Printf("Unexpected server response:\n%v", err)
		os.Exit(1)
	}

	if res.StatusCode != http.StatusOK {
		log.Printf("Unexpected server response:")
		log.Printf("HTTP code: %d", res.StatusCode)
		log.Printf("Response body: %s", body)
		os.Exit(2)
	}

	log.Printf("Server response: %s", body)
}
