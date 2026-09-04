package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"orion-backend/internal/auth"
	"orion-backend/internal/config"
)

func main() {
	nameFlag := flag.String("name", "", "Name of the API key client (e.g. local-client)")
	daysFlag := flag.Int("expires-in-days", 30, "API key lifetime in days (default 30, max 365)")
	storeFlag := flag.String("key-store", "", "Path to key registry JSON file (default from config or data/api_keys.json)")
	flag.Parse()

	if *nameFlag == "" {
		fmt.Fprintln(os.Stderr, "Error: -name flag is required")
		flag.Usage()
		os.Exit(1)
	}

	cfg := config.LoadConfig()

	if err := auth.ValidateSecret(cfg.CorpusAPISecret); err != nil {
		log.Fatalf("Configuration error: %v. Please set CORPUS_API_SECRET (at least 32 characters) in .env", err)
	}

	keyStorePath := cfg.KeyStorePath
	if *storeFlag != "" {
		keyStorePath = *storeFlag
	}

	apiKey, claims, err := auth.GenerateAPIKey(cfg.CorpusAPISecret, *nameFlag, keyStorePath, *daysFlag)
	if err != nil {
		log.Fatalf("Failed to generate API key: %v", err)
	}

	expiresTime := time.Unix(claims.ExpiresAt, 0).UTC().Format(time.RFC3339)

	// Print plaintext key to stdout (for piping / easy capture)
	fmt.Println(apiKey)

	// Print metadata to stderr
	fmt.Fprintf(os.Stderr, "Client Name: %s\n", claims.Name)
	fmt.Fprintf(os.Stderr, "Valid Until: %s\n", expiresTime)
	fmt.Fprintf(os.Stderr, "Key Registry: %s\n", keyStorePath)
}
