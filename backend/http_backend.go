package main

import (
	"fmt"
	"hash/fnv"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"
)

func main() {
	go runHashServer(":80")

	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt)
	<-quit
}

func runHashServer(port string) {
	http.HandleFunc("/hash", hash)
	err := http.ListenAndServe(port, nil)
	if err != nil {
		fmt.Println(err)
	}
}

func hash(responseWriter http.ResponseWriter, request *http.Request) {
	time.Sleep(time.Millisecond * 250)

	requestBody, err := ioutil.ReadAll(request.Body)
	if err != nil {
		log.Printf("Error reading requestBody: %v", err)
		http.Error(responseWriter, "Can't read requestBody", http.StatusBadRequest)
		return
	}

	fnv1a := fnv.New64a()
	fnv1a.Write(requestBody)
	hashedBody := fnv1a.Sum(nil)

	response := append(requestBody, hashedBody...)

	sendHash(responseWriter, response)
}

func sendHash(responseWriter http.ResponseWriter, response []byte) {
	responseWriter.Write(response)
}
