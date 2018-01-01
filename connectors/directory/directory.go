package directory

import (
	"fmt"
	"io"
	"log"
	"mime"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/bugsnag/bugsnag-go/errors"
	"github.com/codem8s/cloud-file-server/config"
	"github.com/codem8s/cloud-file-server/connectors"
)

const (
	prefixURI = "file://"
)

type directoryHandler struct {
	pathPrefix    string
	directoryPath string
}

// New creates a connector that provides files from the directory
func New(config config.ConnectorConfig) (http.Handler, error) {
	if config.URI == "" {
		return nil, errors.Errorf("URI parameter missing in connector %#v", config)
	}
	if !strings.HasPrefix(config.URI, prefixURI) {
		return nil, errors.Errorf("Invalid URI parameter in connector %#v", config)
	}

	directoryPath := strings.Replace(config.URI, prefixURI, "", 1)
	if directoryPath == "" {
		return nil, errors.Errorf("Directory path is missing in URI parameter in connector %#v", config)
	}

	stat, err := os.Stat(directoryPath)
	if os.IsNotExist(err) {
		return nil, errors.Errorf("Directory passed in URI parameter doesn't exists, connector %#v", config)
	}
	if !stat.Mode().IsDir() {
		return nil, errors.Errorf("URI parameter doesn't point to directory, connector %#v", config)
	}

	handler := &directoryHandler{
		pathPrefix:    config.PathPrefix,
		directoryPath: directoryPath,
	}

	return handler, nil
}

func (h *directoryHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	filePath := h.directoryPath + "/" + strings.Replace(req.URL.Path, h.pathPrefix, "", 1)
	filePathFiltered := strings.Replace(filePath, "..", "", -1) // prevents path traversal attack
	log.Printf("Directory request path %q, filePath %q", req.URL.Path, filePathFiltered)

	if strings.HasSuffix(filePath, "/") {
		http.Error(rw, connectors.PageNotFoundMessage, http.StatusNotFound)
		return
	}

	stat, err := os.Stat(filePathFiltered)
	if os.IsNotExist(err) || stat.IsDir() {
		http.Error(rw, connectors.PageNotFoundMessage, http.StatusNotFound)
		return
	}

	ext := path.Ext(filePathFiltered)
	contentType := mime.TypeByExtension(ext)
	if contentType != "" {
		rw.Header().Set("Content-Type", contentType)
	}
	rw.Header().Set("Content-Length", fmt.Sprintf("%d", stat.Size()))

	file, err := os.Open(filePathFiltered)
	if err != nil {
		log.Printf("Error occured during open file %q, %q", filePathFiltered, err)
		http.Error(rw, connectors.InternalServerErrorMessage, http.StatusInternalServerError)
		return
	}

	if _, err := io.Copy(rw, file); err != nil {
		log.Printf("Error occured during copy file %q, %q", filePathFiltered, err)
		http.Error(rw, connectors.InternalServerErrorMessage, http.StatusInternalServerError)
		return
	}
}
