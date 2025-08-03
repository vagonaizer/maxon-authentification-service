package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
	Redis    RedisConfig    `yaml:"redis"`
	JWT      JWTConfig      `yaml:"jwt"`
	Kafka    KafkaConfig    `yaml:"kafka"`
	Logger   LoggerConfig   `yaml:"logger"`
}

type ServerConfig struct {
	HTTPPort        string        `yaml:"http_port" env:"HTTP_PORT"`
	GRPCPort        string        `yaml:"grpc_port" env:"GRPC_PORT"`
	ReadTimeout     time.Duration `yaml:"read_timeout" env:"READ_TIMEOUT"`
	WriteTimeout    time.Duration `yaml:"write_timeout" env:"WRITE_TIMEOUT"`
	ShutdownTimeout time.Duration `yaml:"shutdown_timeout" env:"SHUTDOWN_TIMEOUT"`
	MaxRequestSize  int64         `yaml:"max_request_size" env:"MAX_REQUEST_SIZE"`
	EnableCORS      bool          `yaml:"enable_cors" env:"ENABLE_CORS"`
	EnableRateLimit bool          `yaml:"enable_rate_limit" env:"ENABLE_RATE_LIMIT"`
	RateLimitRPS    int           `yaml:"rate_limit_rps" env:"RATE_LIMIT_RPS"`
}

type DatabaseConfig struct {
	Host            string        `yaml:"host" env:"DB_HOST"`
	Port            string        `yaml:"port" env:"DB_PORT"`
	User            string        `yaml:"user" env:"DB_USER"`
	Password        string        `yaml:"password" env:"DB_PASSWORD"`
	Name            string        `yaml:"name" env:"DB_NAME"`
	SSLMode         string        `yaml:"ssl_mode" env:"DB_SSL_MODE"`
	MaxOpenConns    int           `yaml:"max_open_conns" env:"DB_MAX_OPEN_CONNS"`
	MaxIdleConns    int           `yaml:"max_idle_conns" env:"DB_MAX_IDLE_CONNS"`
	ConnMaxLifetime time.Duration `yaml:"conn_max_lifetime" env:"DB_CONN_MAX_LIFETIME"`
	MigrationsPath  string        `yaml:"migrations_path" env:"DB_MIGRATIONS_PATH"`
}

type RedisConfig struct {
	Host         string        `yaml:"host" env:"REDIS_HOST"`
	Port         string        `yaml:"port" env:"REDIS_PORT"`
	Password     string        `yaml:"password" env:"REDIS_PASSWORD"`
	DB           int           `yaml:"db" env:"REDIS_DB"`
	PoolSize     int           `yaml:"pool_size" env:"REDIS_POOL_SIZE"`
	MinIdleConns int           `yaml:"min_idle_conns" env:"REDIS_MIN_IDLE_CONNS"`
	DialTimeout  time.Duration `yaml:"dial_timeout" env:"REDIS_DIAL_TIMEOUT"`
	ReadTimeout  time.Duration `yaml:"read_timeout" env:"REDIS_READ_TIMEOUT"`
	WriteTimeout time.Duration `yaml:"write_timeout" env:"REDIS_WRITE_TIMEOUT"`
}

type JWTConfig struct {
	AccessTokenSecret  string        `yaml:"access_token_secret" env:"JWT_ACCESS_SECRET"`
	RefreshTokenSecret string        `yaml:"refresh_token_secret" env:"JWT_REFRESH_SECRET"`
	AccessTokenExpiry  time.Duration `yaml:"access_token_expiry" env:"JWT_ACCESS_EXPIRY"`
	RefreshTokenExpiry time.Duration `yaml:"refresh_token_expiry" env:"JWT_REFRESH_EXPIRY"`
	Issuer             string        `yaml:"issuer" env:"JWT_ISSUER"`
	Audience           string        `yaml:"audience" env:"JWT_AUDIENCE"`
}

type KafkaConfig struct {
	Brokers       []string      `yaml:"brokers" env:"KAFKA_BROKERS"`
	GroupID       string        `yaml:"group_id" env:"KAFKA_GROUP_ID"`
	RetryAttempts int           `yaml:"retry_attempts" env:"KAFKA_RETRY_ATTEMPTS"`
	RetryDelay    time.Duration `yaml:"retry_delay" env:"KAFKA_RETRY_DELAY"`
	BatchSize     int           `yaml:"batch_size" env:"KAFKA_BATCH_SIZE"`
	BatchTimeout  time.Duration `yaml:"batch_timeout" env:"KAFKA_BATCH_TIMEOUT"`
}

type LoggerConfig struct {
	Level      string `yaml:"level" env:"LOG_LEVEL"`
	Format     string `yaml:"format" env:"LOG_FORMAT"`
	Output     string `yaml:"output" env:"LOG_OUTPUT"`
	MaxSize    int    `yaml:"max_size" env:"LOG_MAX_SIZE"`
	MaxBackups int    `yaml:"max_backups" env:"LOG_MAX_BACKUPS"`
	MaxAge     int    `yaml:"max_age" env:"LOG_MAX_AGE"`
	Compress   bool   `yaml:"compress" env:"LOG_COMPRESS"`
}

func Load() (*Config, error) {
	cfg := &Config{
		Server: ServerConfig{
			HTTPPort:        getEnv("HTTP_PORT", "8080"),
			GRPCPort:        getEnv("GRPC_PORT", "9090"),
			ReadTimeout:     getDurationEnv("READ_TIMEOUT", 30*time.Second),
			WriteTimeout:    getDurationEnv("WRITE_TIMEOUT", 30*time.Second),
			ShutdownTimeout: getDurationEnv("SHUTDOWN_TIMEOUT", 10*time.Second),
			MaxRequestSize:  getInt64Env("MAX_REQUEST_SIZE", 32<<20),
			EnableCORS:      getBoolEnv("ENABLE_CORS", true),
			EnableRateLimit: getBoolEnv("ENABLE_RATE_LIMIT", true),
			RateLimitRPS:    getIntEnv("RATE_LIMIT_RPS", 100),
		},
		Database: DatabaseConfig{
			Host:            getEnv("DB_HOST", "localhost"),
			Port:            getEnv("DB_PORT", "5432"),
			User:            getEnv("DB_USER", "postgres"),
			Password:        getEnv("DB_PASSWORD", ""),
			Name:            getEnv("DB_NAME", "auth_service"),
			SSLMode:         getEnv("DB_SSL_MODE", "disable"),
			MaxOpenConns:    getIntEnv("DB_MAX_OPEN_CONNS", 25),
			MaxIdleConns:    getIntEnv("DB_MAX_IDLE_CONNS", 5),
			ConnMaxLifetime: getDurationEnv("DB_CONN_MAX_LIFETIME", 5*time.Minute),
			MigrationsPath:  getEnv("DB_MIGRATIONS_PATH", "internal/infrastructure/database/postgres/migrations"),
		},
		Redis: RedisConfig{
			Host:         getEnv("REDIS_HOST", "localhost"),
			Port:         getEnv("REDIS_PORT", "6379"),
			Password:     getEnv("REDIS_PASSWORD", ""),
			DB:           getIntEnv("REDIS_DB", 0),
			PoolSize:     getIntEnv("REDIS_POOL_SIZE", 10),
			MinIdleConns: getIntEnv("REDIS_MIN_IDLE_CONNS", 2),
			DialTimeout:  getDurationEnv("REDIS_DIAL_TIMEOUT", 5*time.Second),
			ReadTimeout:  getDurationEnv("REDIS_READ_TIMEOUT", 3*time.Second),
			WriteTimeout: getDurationEnv("REDIS_WRITE_TIMEOUT", 3*time.Second),
		},
		JWT: JWTConfig{
			AccessTokenSecret:  getEnv("JWT_ACCESS_SECRET", ""),
			RefreshTokenSecret: getEnv("JWT_REFRESH_SECRET", ""),
			AccessTokenExpiry:  getDurationEnv("JWT_ACCESS_EXPIRY", 15*time.Minute),
			RefreshTokenExpiry: getDurationEnv("JWT_REFRESH_EXPIRY", 24*time.Hour*7),
			Issuer:             getEnv("JWT_ISSUER", "auth-service"),
			Audience:           getEnv("JWT_AUDIENCE", "social-network"),
		},
		Kafka: KafkaConfig{
			Brokers:       getSliceEnv("KAFKA_BROKERS", []string{"localhost:9092"}),
			GroupID:       getEnv("KAFKA_GROUP_ID", "auth-service"),
			RetryAttempts: getIntEnv("KAFKA_RETRY_ATTEMPTS", 3),
			RetryDelay:    getDurationEnv("KAFKA_RETRY_DELAY", 1*time.Second),
			BatchSize:     getIntEnv("KAFKA_BATCH_SIZE", 100),
			BatchTimeout:  getDurationEnv("KAFKA_BATCH_TIMEOUT", 1*time.Second),
		},
		Logger: LoggerConfig{
			Level:      getEnv("LOG_LEVEL", "info"),
			Format:     getEnv("LOG_FORMAT", "json"),
			Output:     getEnv("LOG_OUTPUT", "stdout"),
			MaxSize:    getIntEnv("LOG_MAX_SIZE", 100),
			MaxBackups: getIntEnv("LOG_MAX_BACKUPS", 3),
			MaxAge:     getIntEnv("LOG_MAX_AGE", 28),
			Compress:   getBoolEnv("LOG_COMPRESS", true),
		},
	}

	return cfg, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getIntEnv(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getInt64Env(key string, defaultValue int64) int64 {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.ParseInt(value, 10, 64); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getBoolEnv(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

func getDurationEnv(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}

func getSliceEnv(key string, defaultValue []string) []string {
	if value := os.Getenv(key); value != "" {
		return []string{value}
	}
	return defaultValue
}
