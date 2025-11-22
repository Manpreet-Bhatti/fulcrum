package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
)

type Config struct {
	LBPort   int      `json:"lb_port"`
	Backends []string `json:"backends"`
}

func LoadConfig(file string) (*Config, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var config Config
	decoder := json.NewDecoder(f)
	err = decoder.Decode(&config)
	return &config, err
}

func main() {
	config, err := LoadConfig("config.json")
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	fmt.Printf("⚖️  Fulcrum Load Balancer starting on port %d\n", config.LBPort)

	targetURL := config.Backends[0]
	parsedURL, err := url.Parse(targetURL)
	if err != nil {
		log.Fatalf("Invalid backend URL: %v", err)
	}

	proxy := httputil.NewSingleHostReverseProxy(parsedURL)
	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)
		req.Host = parsedURL.Host
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("[Fulcrum] Forwarding request -> %s\n", targetURL)

		proxy.ServeHTTP(w, r)
	})

	listenAddr := fmt.Sprintf(":%d", config.LBPort)
	if err := http.ListenAndServe(listenAddr, nil); err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}
