# deployment file for the go-multitenancy framework
FROM golang:latest

WORKDIR /go/src/app
COPY . .

RUN go get -d -v ./...
RUN go build
RUN ./home-crafting-backend

MAINTAINER Liam Read