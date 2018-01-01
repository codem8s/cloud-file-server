FROM golang:1.9.2-alpine

RUN apk add --no-cache --update alpine-sdk

COPY . /go/src/github.com/codem8s/cloud-file-server
RUN cd /go/src/github.com/codem8s/cloud-file-server && go build

FROM alpine:3.6

EXPOSE 8080

RUN apk --no-cache upgrade && \
    apk --no-cache add --update ca-certificates bash

COPY --from=0 /go/src/github.com/codem8s/cloud-file-server/cloud-file-server /usr/local/bin/cloud-file-server

ENTRYPOINT ["cloud-file-server"]