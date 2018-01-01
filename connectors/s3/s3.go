package s3

import (
	"fmt"
	"io"
	"log"
	"mime"
	"net/http"
	"path"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/bugsnag/bugsnag-go/errors"
	"github.com/codem8s/cloud-file-server/config"
	"github.com/codem8s/cloud-file-server/connectors"
)

const (
	prefixURI = "s3://"
)

type s3Handler struct {
	pathPrefix    string
	bucketName    string
	bucketFolders string
	svc           *s3.S3
}

// New creates a connector that provides files from AWS s3 bucket
func New(config config.ConnectorConfig) (http.Handler, error) {
	if config.URI == "" {
		return nil, errors.Errorf("URI parameter missing in connector %#v", config)
	}
	if config.Region == "" {
		return nil, errors.Errorf("Region parameter missing in connector %#v", config)
	}
	if !strings.HasPrefix(config.URI, prefixURI) {
		return nil, errors.Errorf("Invalid URI parameter in connector %#v", config)
	}

	uriWithOutS3Prefix := strings.Replace(config.URI, prefixURI, "", 1)
	if uriWithOutS3Prefix == "" {
		return nil, errors.Errorf("Bucket name missing in URI parameter in connector %#v", config)
	}

	var bucketName string
	bucketFolders := ""
	if index := strings.Index(uriWithOutS3Prefix, "/"); index == -1 {
		bucketName = uriWithOutS3Prefix
	} else {
		bucketName = uriWithOutS3Prefix[:index]
		bucketFolders = uriWithOutS3Prefix[index:]
	}

	svc := s3.New(session.New(), &aws.Config{
		Region: &config.Region,
	})

	handler := &s3Handler{
		pathPrefix:    config.PathPrefix,
		bucketName:    bucketName,
		bucketFolders: bucketFolders,
		svc:           svc,
	}

	return handler, nil
}

func (h *s3Handler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	key := h.bucketFolders + strings.Replace(req.URL.Path, h.pathPrefix, "", 1)
	log.Printf("S3 request path %q, key %q", req.URL.Path, key)

	if strings.HasSuffix(key, "/") {
		http.Error(rw, connectors.PageNotFoundMessage, http.StatusNotFound)
		return
	}

	input := &s3.GetObjectInput{
		Bucket: aws.String(h.bucketName),
		Key:    aws.String(key),
	}
	if v := req.Header.Get("If-None-Match"); v != "" {
		input.IfNoneMatch = aws.String(v)
	}

	var is304 bool
	resp, err := h.svc.GetObject(input)
	if awsErr, ok := err.(awserr.Error); ok {
		switch awsErr.Code() {
		case s3.ErrCodeNoSuchKey:
			http.Error(rw, connectors.PageNotFoundMessage, http.StatusNotFound)
			return
		case "NotModified":
			is304 = true
		// continue so other headers get set appropriately
		default:
			log.Printf("Error: %v %v", awsErr.Code(), awsErr.Message())
			http.Error(rw, connectors.InternalServerErrorMessage, http.StatusInternalServerError)
			return
		}
	} else if err != nil {
		log.Printf("not aws error %v %s", err, err)
		http.Error(rw, connectors.InternalServerErrorMessage, http.StatusInternalServerError)
		return
	}

	var contentType string
	if resp.ContentType != nil {
		contentType = *resp.ContentType
	}

	if contentType == "" {
		ext := path.Ext(req.URL.Path)
		contentType = mime.TypeByExtension(ext)
	}

	if resp.ETag != nil && *resp.ETag != "" {
		rw.Header().Set("Etag", *resp.ETag)
	}

	if contentType != "" {
		rw.Header().Set("Content-Type", contentType)
	}
	if resp.ContentLength != nil && *resp.ContentLength > 0 {
		rw.Header().Set("Content-Length", fmt.Sprintf("%d", *resp.ContentLength))
	}

	if is304 {
		rw.WriteHeader(304)
	} else {
		_, err := io.Copy(rw, resp.Body)
		if err != nil {
			log.Printf("Error occured during copy file %q, %q", key, err)
			http.Error(rw, connectors.InternalServerErrorMessage, http.StatusInternalServerError)
			return
		}
		resp.Body.Close()
	}
}
