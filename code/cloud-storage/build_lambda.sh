#!/bin/bash

GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o build/main ./cmd/lambda/main.go
zip -jrm build/main.zip build/main

aws lambda update-function-code --function-name get-s3-object --zip-file fileb://build/main.zip
