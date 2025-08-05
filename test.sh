#!/bin/bash

# cd src && go test -race ./... 
# cd src && go test -coverprofile=coverage.out ./... && go tool cover -func=coverage.out

cd src && go test -coverprofile=coverage.out ./app/features/...

go tool cover -html=coverage.out -o coverage.html