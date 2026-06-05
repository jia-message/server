package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	Port         string
	Env          string
	MasterKey    string
	JWTSecret    string
	DBHost       string
	DBPort       string
	DBUser       string
	DBPassword   string
	DBName       string
	DBSSLMode    string
}

var AppConfig *Config

func LoadConfig() {
	// Load .env if it exists
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, relying on system environment variables")
	}

	masterKey := os.Getenv("JIA_MASTER_KEY")
	if masterKey == "" {
		log.Fatal("JIA_MASTER_KEY environment variable is mandatory for encrypting credentials at rest")
	}
	if len(masterKey) < 16 {
		log.Fatal("JIA_MASTER_KEY must be at least 16 bytes for security")
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "default-change-me-in-production-jwt-secret-key-12345"
		log.Println("WARNING: JWT_SECRET not set, using insecure default!")
	}

	AppConfig = &Config{
		Port:       getEnv("PORT", "3000"),
		Env:        getEnv("ENV", "development"),
		MasterKey:  masterKey,
		JWTSecret:  jwtSecret,
		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "5432"),
		DBUser:     getEnv("DB_USER", "postgres"),
		DBPassword: getEnv("DB_PASSWORD", "postgres"),
		DBName:     getEnv("DB_NAME", "jia"),
		DBSSLMode:  getEnv("DB_SSLMODE", "disable"),
	}
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	str := getEnv(key, "")
	if value, err := strconv.Atoi(str); err == nil {
		return value
	}
	return defaultValue
}
