package handlers

import (
	"log"
	"net/http"
	"strings"

	"github.com/codem8s/cloud-file-server/config"
	"github.com/codem8s/cloud-file-server/connectors"
	"github.com/codem8s/cloud-file-server/connectors/directory"
	"github.com/codem8s/cloud-file-server/connectors/file"
	"github.com/codem8s/cloud-file-server/connectors/s3"
)

const (
	s3Connector        = "s3"
	fileConnector      = "file"
	directoryConnector = "directory"
)

// MainHandler describes main handler configuration
type MainHandler struct {
	connectors map[string]http.Handler
}

func (h *MainHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	path := req.URL.Path
	for pathPrefix, connector := range h.connectors {
		if strings.HasPrefix(path, pathPrefix) {
			connector.ServeHTTP(rw, req)
			return
		}
	}

	http.Error(rw, connectors.PageNotFoundMessage, http.StatusNotFound)
}

// New creates a main http handler which uses provided connectors
func New(config config.Config) http.Handler {
	if len(config.Connectors) <= 0 {
		log.Fatal("There are no connectors in config")
	}

	var conns = make(map[string]http.Handler)
	for _, connectorConfig := range config.Connectors {
		if connectorConfig.PathPrefix == "" {
			log.Fatalf("PathPrefix is missing in connector config %#v", connectorConfig)
		}

		var connectorHandler http.Handler
		var err error

		switch connectorConfig.Type {
		case s3Connector:
			connectorHandler, err = s3.New(connectorConfig)
		case directoryConnector:
			connectorHandler, err = directory.New(connectorConfig)
		case fileConnector:
			connectorHandler, err = file.New(connectorConfig)
		default:
			log.Fatalf("Invalid type of connector config %#v", connectorConfig)
		}
		if err != nil {
			log.Fatal(err)
		}
		if conns[connectorConfig.PathPrefix] != nil {
			log.Fatalf("Found duplicated pathPrefix %#v", connectorConfig)
		}
		conns[connectorConfig.PathPrefix] = connectorHandler
		log.Printf("Created handler %#v", connectorHandler)
	}

	return &MainHandler{connectors: conns}
}
