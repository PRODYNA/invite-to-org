FROM golang:1.22.1-alpine3.19 as build

WORKDIR /app
COPY . /app
RUN GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build --ldflags '-extldflags=-static' -o invite-to-org main.go

FROM alpine:3.19.1
COPY --from=build /app/invite-to-org /app/
ENTRYPOINT ["/app/invite-to-org"]
