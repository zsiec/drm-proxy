package main

import (
	"log"
	"net/http"
	"time"

	"github.com/cbsinteractive/drm-proxy/pkg/drmproxy/providers/irdeto"
	"github.com/cbsinteractive/drm-proxy/pkg/restapi"
	"github.com/kelseyhightower/envconfig"
)

type config struct {
	Endpoint string `envconfig:"IRDETO_ENDPOINT"`
	Token    string `envconfig:"IRDETO_AUTH_TOKEN"`
}

func main() {
	client := http.Client{
		Timeout: 10 * time.Second,
	}

	cfg, err := loadConfig()
	if err != nil {
		panic(err)
	}

	contentKeySvc := irdeto.NewKeyService(cfg.Endpoint, cfg.Token, client)

	timeout := 10 * time.Second
	port := 7777
	server, err := restapi.NewServer(timeout, port, contentKeySvc)
	if err != nil {
		panic(err)
	}

	log.Printf("Server starting and listening on port %d", port)

	panic(server.ListenAndServe())
}

// loadConfig loads the configuration from environment variables.
func loadConfig() (config, error) {
	var c config
	err := envconfig.Process("drm_proxy", &c)
	return c, err
}
