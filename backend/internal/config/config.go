package config

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	DBHost             string
	DBPort             string
	DBUser             string
	DBPassword         string
	DBName             string
	RedisHost          string
	RedisPort          string
	MinioEndpoint      string
	MinioAccessKey     string
	MinioSecretKey     string
	MinioBucket        string
	MinioUseSSL        bool
	NatsURL            string
	JWTSecret          string
	JWTExpirationHours int
	ServerPort         string
	ServerHost         string
	PromAPIBaseURL     string
	PromAPIKey         string
	EncryptionKey      string
}

func Load() (*Config, error) {
	godotenv.Load("../.env") // load from project root
	godotenv.Load(".env")    // or from backend dir

	expHours, _ := strconv.Atoi(getEnv("JWT_EXPIRATION_HOURS", "72"))
	useSSL, _ := strconv.ParseBool(getEnv("MINIO_USE_SSL", "false"))

	return &Config{
		DBHost:             getEnv("DB_HOST", "localhost"),
		DBPort:             getEnv("DB_PORT", "5432"),
		DBUser:             getEnv("DB_USER", "sellflow"),
		DBPassword:         getEnv("DB_PASSWORD", "sellflow_secret"),
		DBName:             getEnv("DB_NAME", "sellflow"),
		RedisHost:          getEnv("REDIS_HOST", "localhost"),
		RedisPort:          getEnv("REDIS_PORT", "6379"),
		MinioEndpoint:      getEnv("MINIO_ENDPOINT", "localhost:9000"),
		MinioAccessKey:     getEnv("MINIO_ACCESS_KEY", "minioadmin"),
		MinioSecretKey:     getEnv("MINIO_SECRET_KEY", "minioadmin"),
		MinioBucket:        getEnv("MINIO_BUCKET", "sellflow"),
		MinioUseSSL:        useSSL,
		NatsURL:            getEnv("NATS_URL", "nats://localhost:4222"),
		JWTSecret:          getEnv("JWT_SECRET", "change-me"),
		JWTExpirationHours: expHours,
		ServerPort:         getEnv("SERVER_PORT", "8080"),
		ServerHost:         getEnv("SERVER_HOST", "0.0.0.0"),
		PromAPIBaseURL:     getEnv("PROM_API_BASE_URL", "https://my.prom.ua/api/v1"),
		PromAPIKey:         getEnv("PROM_API_KEY", ""),
		EncryptionKey:      getEnv("ENCRYPTION_KEY", ""),
	}, nil
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func (c *Config) DSN() string {
	return "host=" + c.DBHost + " user=" + c.DBUser + " password=" + c.DBPassword + " dbname=" + c.DBName + " port=" + c.DBPort + " sslmode=disable TimeZone=Europe/Kyiv"
}
