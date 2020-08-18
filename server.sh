#!/bin/bash

go get ./...

go build server/udp_server.go

./udp_server $1 $2