package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	HTTPAddr                      string
	LogEnabled                    bool
	ProfilingEnabled              bool
	ProfilingAddr                 string
	ProfilingRuntimeStatsEnabled  bool
	ProfilingRuntimeStatsInterval time.Duration
	DatabaseURL                   string
	RedisEnabled                  bool
	RedisAddr                     string
	RedisPassword                 string
	RedisDB                       int
	RedisTTL                      time.Duration
	RedisDialTO                   time.Duration
	RedisReadTO                   time.Duration
	RedisWriteTO                  time.Duration
	CBMaxRequests                 uint32
	CBInterval                    time.Duration
	CBTimeout                     time.Duration
	CBFailureRatio                float64
	CBMinRequests                 uint32
	RequestTimeout                time.Duration
	ReadTimeout                   time.Duration
	WriteTimeout                  time.Duration
	IdleTimeout                   time.Duration
	CORSAllowOrigin               string
	MaxConns                      int32
	MinConns                      int32
}

func Load() Config {
	return Config{
		HTTPAddr:                      getString("HTTP_ADDR", ":8000"),
		LogEnabled:                    getBool("LOG_ENABLED", true),
		ProfilingEnabled:              getBool("PROFILING_ENABLED", false),
		ProfilingAddr:                 getString("PROFILING_ADDR", "127.0.0.1:6060"),
		ProfilingRuntimeStatsEnabled:  getBool("PROFILING_RUNTIME_STATS_ENABLED", true),
		ProfilingRuntimeStatsInterval: getDuration("PROFILING_RUNTIME_STATS_INTERVAL", 15*time.Second),
		DatabaseURL:                   getString("DATABASE_URL", "postgresql://postgres:postgres@localhost:5432/url_shortener?sslmode=disable"),
		RedisEnabled:                  getBool("REDIS_ENABLED", false),
		RedisAddr:                     getString("REDIS_ADDR", "localhost:6379"),
		RedisPassword:                 getString("REDIS_PASSWORD", ""),
		RedisDB:                       getInt("REDIS_DB", 0),
		RedisTTL:                      getDuration("REDIS_TTL", 24*time.Hour),
		RedisDialTO:                   getDuration("REDIS_DIAL_TIMEOUT", 2*time.Second),
		RedisReadTO:                   getDuration("REDIS_READ_TIMEOUT", 500*time.Millisecond),
		RedisWriteTO:                  getDuration("REDIS_WRITE_TIMEOUT", 500*time.Millisecond),
		CBMaxRequests:                 getUint32("CB_MAX_REQUESTS", 3),
		CBInterval:                    getDuration("CB_INTERVAL", 30*time.Second),
		CBTimeout:                     getDuration("CB_TIMEOUT", 10*time.Second),
		CBFailureRatio:                getFloat64("CB_FAILURE_RATIO", 0.6),
		CBMinRequests:                 getUint32("CB_MIN_REQUESTS", 5),
		RequestTimeout:                getDuration("REQUEST_TIMEOUT", 2*time.Second),
		ReadTimeout:                   getDuration("READ_TIMEOUT", 15*time.Second),
		WriteTimeout:                  getDuration("WRITE_TIMEOUT", 15*time.Second),
		IdleTimeout:                   getDuration("IDLE_TIMEOUT", 60*time.Second),
		CORSAllowOrigin:               getString("CORS_ALLOW_ORIGIN", "*"),
		MaxConns:                      getInt32("DB_MAX_CONNS", 10),
		MinConns:                      getInt32("DB_MIN_CONNS", 2),
	}
}

func getString(key string, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}

func getDuration(key string, fallback time.Duration) time.Duration {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	parsed, err := time.ParseDuration(value)
	if err != nil {
		return fallback
	}

	return parsed
}

func getInt32(key string, fallback int32) int32 {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	parsed, err := strconv.ParseInt(value, 10, 32)
	if err != nil {
		return fallback
	}

	return int32(parsed)
}

func getInt(key string, fallback int) int {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}

	return parsed
}

func getUint32(key string, fallback uint32) uint32 {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	parsed, err := strconv.ParseUint(value, 10, 32)
	if err != nil {
		return fallback
	}

	return uint32(parsed)
}

func getFloat64(key string, fallback float64) float64 {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	parsed, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return fallback
	}

	return parsed
}

func getBool(key string, fallback bool) bool {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return fallback
	}

	return parsed
}
