package config

import (
	"flag"
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	HTTPServer HTTPServer `json:"api"`
	NarWAL     NarWAL     `json:"narwal"`
	Debug      bool       `json:"debug"`
}

type HTTPServer struct {
	Host string `json:"host"`
	Port string `json:"port"`
}

// Address for binding web server
func (s HTTPServer) Address() string {
	return fmt.Sprintf("%s:%s", s.Host, s.Port)
}

// NarWAL keeps config of ni-storage
type NarWAL struct {
	DataDir string `json:"data-dir"`
}

// Load config from environment and command line
func Load() *Config {
	c := &Config{
		HTTPServer: HTTPServer{
			Host: "0.0.0.0",
			Port: "8555",
		},
		NarWAL: NarWAL{},
	}
	c.loadFromEnv()
	c.loadFromCLI()
	return c
}

func (c *Config) loadFromEnv() {
	if v := os.Getenv("NI_API_HOST"); v != "" {
		c.HTTPServer.Host = v
	}
	if v := os.Getenv("NI_API_PORT"); v != "" {
		c.HTTPServer.Port = v
	}
	if v := os.Getenv("NI_NARWAL_DATA_DIR"); v != "" {
		c.NarWAL.DataDir = v
	}
	if v := os.Getenv("NI_DEBUG"); v == "true" {
		c.Debug = true
	}
}

func (c *Config) loadFromCLI() {
	var (
		host    string
		port    int
		dataDir string
		debug   bool
	)
	flag.StringVar(&host, "host", "", "api-server host (default: 0.0.0.0)")
	flag.IntVar(&port, "port", 0, "api-server port (default: 8500)")
	flag.StringVar(&dataDir, "data-dir", "./data", "path to folder with data")
	flag.BoolVar(&debug, "debug", false, "debug mode with verbose logging")
	flag.Parse()

	if host != "" {
		c.HTTPServer.Host = host
	}
	if port > 0 {
		c.HTTPServer.Port = strconv.Itoa(port)
	}
	if dataDir != "" {
		c.NarWAL.DataDir = dataDir
	}
	if debug {
		c.Debug = true
	}
}
