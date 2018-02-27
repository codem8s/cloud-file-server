# cloud-file-server

cloud-file-server is a an application which provide files via http using configured connectors

[![Build Status](https://travis-ci.org/codem8s/cloud-file-server.svg?branch=master)](https://travis-ci.org/codem8s/cloud-file-server)
[![Go Report Card](https://goreportcard.com/badge/github.com/codem8s/cloud-file-server)](https://goreportcard.com/report/github.com/codem8s/cloud-file-server)
[![GoDoc](https://godoc.org/github.com/codem8s/cloud-file-server?status.svg "GoDoc Documentation")](https://godoc.org/github.com/codem8s/cloud-file-server)

## Connectors
- AWS s3 bucket
- local directory
- local single file

## Run
    ./cloud-file-server --config example-config.yaml
    
## Example config

    listen: :8080
    logRequests: true
    connectors:
    - type: s3
      uri: s3://aws-s3-bucket-name/example/path
      region: eu-west-1
      pathPrefix: /s3
    - type: file
      uri: file:///example/path/file.yaml
      pathPrefix: /file
    - type: directory
      uri: file:///example/path/directory
      pathPrefix: /dir
