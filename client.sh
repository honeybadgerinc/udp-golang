#!/bin/bash

go get ./...

go build client/udp_client.go

./udp_client $1 $2