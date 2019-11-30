# go-update-route53 
A Go program to update AWS Route53 with a record pointing to the public IP where the program is running. 

## Setup

1. Build go-update-route53 docker image:
```
docker build -t go-update-route53 .
```

## Run

`go run cmd/main.go -host=localhost -port=8080 -id=XXXXXXXXXXXX -file=output.csv`

