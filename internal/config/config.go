package config

import (
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	ServerName       string
	ServerPort       string
	ServerEnv        string
	MongoDBURI       string
	CorpusAPISecret  string
	KeyStorePath     string
	CORSAllowOrigins []string
}

// LoadConfig reads configuration from .env file or environment variables.
func LoadConfig() *Config {
	env := strings.ToLower(strings.TrimSpace(os.Getenv("SERVER_ENV")))
	if env == "production" || env == "prod" {
		log.Println("Running in production mode: skipping .env file loading")
	} else {
		if err := godotenv.Load(); err != nil {
			log.Println("Note: .env file not found or failed to load, reading from environment variables")
		}
	}

	rawCORS := getEnv("CORS_ALLOW_ORIGINS", "*")
	var origins []string
	for _, origin := range strings.Split(rawCORS, ",") {
		trimmed := strings.TrimSpace(origin)
		if trimmed != "" {
			origins = append(origins, trimmed)
		}
	}
	if len(origins) == 0 {
		origins = []string{"*"}
	}

	return &Config{
		ServerName:       getEnv("SERVER_NAME", "orion-backend"),
		ServerPort:       getEnv("SERVER_PORT", "8080"),
		ServerEnv:        getEnv("SERVER_ENV", "development"),
		MongoDBURI:       getEnv("MONGODB_URI", "mongodb://admin:166333@localhost:27017/orion_db?authSource=admin"),
		CorpusAPISecret:  getEnv("CORPUS_API_SECRET", ""),
		KeyStorePath:     getEnv("API_KEY_STORE_PATH", "data/api_keys.json"),
		CORSAllowOrigins: origins,
	}
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}
