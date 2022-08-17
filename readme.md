# Hatchways test
This is a GO project in which I wrote a simple backend JSON Restful API. The purpose of the API is to return JSON formatted data containing blog posts with the specified tags. The blog posts are retrieved from an external api.

## Requirements
GO 1.17+

## Run
go run main.go

## Installation
go build

## Testing
go test -v

## Usage
1. Run the following command: go run main.go
2. API is listening on 0.0.0.0:8000 (on windows localhost:8000)
3. Example query: http://localhost:8080/api/posts?tags=science

## License
MIT License, check LICENSE
