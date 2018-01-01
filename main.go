package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"flag"
	"io/ioutil"

	"github.com/codem8s/cloud-file-server/config"
	mainhandler "github.com/codem8s/cloud-file-server/handlers"
	"github.com/gorilla/handlers"
	"gopkg.in/yaml.v2"
)

func main() {
	configFile := flag.String("config", "", "Configuration file path")
	flag.Parse()

	if *configFile == "" {
		log.Fatal("'--config' parameter is missing")
	}

	data, err := ioutil.ReadFile(*configFile)
	if err != nil {
		log.Fatalf("Can not read config %s", err.Error())
	}

	var cfg config.Config
	if err = yaml.UnmarshalStrict(data, &cfg); err != nil {
		log.Fatalf("Can not parse config %s", err.Error())
	}

	if cfg.Listen == "" {
		cfg.Listen = ":8080"
	}

	handler := mainhandler.New(cfg)
	if cfg.LogRequests {
		handler = handlers.CombinedLoggingHandler(os.Stdout, handler)
	}

	server := &http.Server{
		Addr:           cfg.Listen,
		Handler:        handler,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	log.Printf("Listening on %s", cfg.Listen)
	log.Fatal(server.ListenAndServe())
}
