package file

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

type fileHandler struct {
	pathPrefix string
	filePath   string
	fileName   string
}

// New creates a connector that provides the file from the directory
func New(config config.ConnectorConfig) (http.Handler, error) {
	if config.URI == "" {
		return nil, errors.Errorf("URI parameter missing in connector %#v", config)
	}
	if !strings.HasPrefix(config.URI, prefixURI) {
		return nil, errors.Errorf("Invalid URI parameter in connector %#v", config)
	}

	filePath := strings.Replace(config.URI, prefixURI, "", 1)
	if filePath == "" {
		return nil, errors.Errorf("File path is missing in URI parameter in connector %#v", config)
	}

	stat, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		return nil, errors.Errorf("File passed in URI parameter doesn't exists, connector %#v", config)
	}
	if stat.Mode().IsDir() {
		return nil, errors.Errorf("URI parameter doesn't point to file, connector %#v", config)
	}

	handler := &fileHandler{
		pathPrefix: config.PathPrefix,
		filePath:   filePath,
		fileName:   path.Base(filePath),
	}

	return handler, nil
}

func (h *fileHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	filePath := strings.Replace(req.URL.Path, h.pathPrefix, "", 1)
	log.Printf("Directory request path %q, filePath %q", req.URL.Path, filePath)

	if path.Base(filePath) != h.fileName || (len(filePath)-1) != len(h.fileName) {
		http.Error(rw, connectors.PageNotFoundMessage, http.StatusNotFound)
		return
	}

	stat, err := os.Stat(h.filePath)
	if os.IsNotExist(err) {
		http.Error(rw, connectors.PageNotFoundMessage, http.StatusNotFound)
		return
	}

	ext := path.Ext(h.filePath)
	contentType := mime.TypeByExtension(ext)
	if contentType != "" {
		rw.Header().Set("Content-Type", contentType)
	}
	rw.Header().Set("Content-Length", fmt.Sprintf("%d", stat.Size()))

	file, err := os.Open(h.filePath)
	if err != nil {
		log.Printf("Error occured during open file %q, %q", h.filePath, err)
		http.Error(rw, connectors.InternalServerErrorMessage, http.StatusInternalServerError)
		return
	}

	if _, err := io.Copy(rw, file); err != nil {
		log.Printf("Error occured during copy file %q, %q", h.filePath, err)
		http.Error(rw, connectors.InternalServerErrorMessage, http.StatusInternalServerError)
		return
	}
}
