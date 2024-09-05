package main

import (
	"flag"
	"os"
	"path/filepath"
)

type Config struct {
	Addr            string // :port or addr:port to listen on
	CredentialsFile string // path to service-account.json file
}

func parseCommandLine() *Config {
	// addr
	dfltAddr := "localhost:8842"
	addr := flag.String("addr", dfltAddr, "Address to listen on")

	// credentialsFile defaults to env var, then relative path, then filename
	dfltCredentialsFile := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")
	if dfltCredentialsFile == "" {
		exe, err := os.Executable()
		if err != nil {
			dfltCredentialsFile = "service-account.json"
		} else {
			dfltCredentialsFile = filepath.Join(filepath.Dir(exe), "service-account.json")
		}
	}
	credentialsFile := flag.String("credentialsFile", dfltCredentialsFile, "Path to service account credentials in JSON format")

	flag.Parse()

	return &Config{
		Addr:            *addr,
		CredentialsFile: *credentialsFile,
	}
}
