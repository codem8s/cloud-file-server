# cloud-file-server

cloud-file-server is a an application which provide files via http using configured connectors

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