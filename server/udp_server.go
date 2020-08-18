package main

import (
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
)

func main() {

	UDPAddress := os.Args[1]
	HTTPAddress := os.Args[2]

	go runUdpServer(UDPAddress, HTTPAddress)

	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt)
	<-quit

}

func runUdpServer(address string, httpAddress string) {
	network := "udp4"
	var addressWithPort strings.Builder
	addressWithPort.WriteString(address)
	addressWithPort.WriteString(":40000")
	udpAddress, err := net.ResolveUDPAddr(network, addressWithPort.String())
	if err != nil {
		fmt.Println(err)
		return
	}

	connection, err := net.ListenUDP(network, udpAddress)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer connection.Close()

	maxBufferSize := 10000
	readInBuffer := make([]byte, maxBufferSize)
	var receivedPacketCount int
	go func() {
		for {
			bytesReadIn, _, err := connection.ReadFromUDP(readInBuffer)
			fmt.Println("bytes read in -> ", bytesReadIn)
			var httpIPString strings.Builder
			httpIPString.WriteString(httpAddress)
			httpIPString.WriteString(":80/hash")
			request, _ := http.NewRequest("GET", httpIPString.String(), nil)
			query := request.URL.Query()
			query.Add("data", string(readInBuffer))
			request.URL.RawQuery = query.Encode()
			response, err := http.Get(request.URL.String())
			receivedPacketCount++
			responseBody, _ := ioutil.ReadAll(response.Body)
			fmt.Println("bytes read in -> ", len(responseBody))
			_, err = connection.WriteToUDP(readInBuffer, &net.UDPAddr{
				IP:   net.ParseIP(address),
				Port: 40001,
				Zone: "",
			})
			readInBuffer = make([]byte, maxBufferSize)
			if err != nil {
				fmt.Println(err)
				return
			}
			fmt.Println(receivedPacketCount)
		}
	}()

	return
}
