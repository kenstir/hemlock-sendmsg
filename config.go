package main

import (
	"flag"
)

type Config struct {
	Addr string // :port or addr:port to listen on
	CredentialsFile string // path to service-account.json file
}

func parseCommandLine() *Config {
	dfltAddr := "localhost:8842"
	addr := flag.String("addr", dfltAddr, "Address to listen on")
	dfltCredentialsFile := "service-account.json"
	credentialsFile := flag.String("credentialsFile", dfltCredentialsFile, "Path to service account credentials in JSON format")

	flag.Parse()

	return &Config{
		Addr:            *addr,
		CredentialsFile: *credentialsFile,
	}
}
