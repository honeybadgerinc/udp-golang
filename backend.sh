#!/bin/bash

go get ./...

go build backend/http_backend.go

./http_backend