package config

import (
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

type RedisConfig struct {
	Host         string `mapstructure:"host" validate:"required"`
	Port         int    `mapstructure:"port" validate:"required"`
	Password     string `mapstructure:"password" validate:"required"`
	DB           int    `mapstructure:"db" validate:"required"`
	PoolSize     int    `mapstructure:"pool_size" validate:"required"`
	MinIdleConns int    `mapstructure:"min_idle_conns" validate:"required"`
	DialTimeout  int    `mapstructure:"dial_timeout" validate:"required"`
	ReadTimeout  int    `mapstructure:"read_timeout" validate:"required"`
	WriteTimeout int    `mapstructure:"write_timeout" validate:"required"`
	MaxRetries   int    `mapstructure:"max_retries" validate:"required"`
}

type ServerConfig struct {
	Host        string `mapstructure:"host" validate:"required"`
	Port        int    `mapstructure:"port" validate:"required,min=1024,max=65535"`
	Mode        string `mapstructure:"mode" validate:"required"`
	DocsEnabled bool   `mapstructure:"docs_enabled" validate:"required"`
}

type AuthConfig struct {
	JwtSecret     string `mapstructure:"jwt_secret" validate:"required"`
	JwtAccessTTL  int    `mapstructure:"jwt_expire_time" validate:"required"`         // seconds
	JwtRefreshTTL int    `mapstructure:"jwt_refresh_expire_time" validate:"required"` // seconds
}

type MemoryCacheConfig struct {
	CleanupInterval int `mapstructure:"cleanup_interval" validate:"required"`
	MaxItems        int `mapstructure:"max_items" validate:"required"`
}

type CacheConfig struct {
	Type       string            `mapstructure:"type" validate:"required"`
	Prefix     string            `mapstructure:"prefix" validate:"required"`
	DefaultTTL int               `mapstructure:"default_ttl" validate:"required"`
	Redis      RedisConfig       `mapstructure:"redis" validate:"required"`
	Memory     MemoryCacheConfig `mapstructure:"memory" validate:"required"`
}

type DatabaseLoggingConfig struct {
	Level            string `mapstructure:"level" validate:"required"`
	FilePath         string `mapstructure:"file_path" validate:"required"`
	MaxSizeMB        int    `mapstructure:"max_size_mb" validate:"required"`
	MaxBackups       int    `mapstructure:"max_backups" validate:"required"`
	MaxAgeDays       int    `mapstructure:"max_age_days" validate:"required"`
	Compress         bool   `mapstructure:"compress" validate:"required"`
	Console          bool   `mapstructure:"console" validate:"required"`
	SlowSqlThreshold int    `mapstructure:"slow_sql_threshold" validate:"required"` // seconds
}

type DatabaseConfig struct {
	Driver          string                `mapstructure:"driver" validate:"required"` // mysql, postgres, sqlite, sqlserver
	Host            string                `mapstructure:"host" validate:"required"`
	Port            int                   `mapstructure:"port" validate:"required"`
	User            string                `mapstructure:"user" validate:"required"`
	Password        string                `mapstructure:"password" validate:"required"`
	DBName          string                `mapstructure:"dbname" validate:"required"`
	SSLMode         string                `mapstructure:"sslmode" validate:"required"`
	MaxOpenConns    int                   `mapstructure:"max_open_conns" validate:"required"`
	MaxIdleConns    int                   `mapstructure:"max_idle_conns" validate:"required"`
	ConnMaxLifetime int                   `mapstructure:"conn_max_lifetime" validate:"required"` // seconds
	Logging         DatabaseLoggingConfig `mapstructure:"logging" validate:"required"`
}

type EmailConfig struct {
	Provider string     `mapstructure:"provider" validate:"required"`
	SMTP     SMTPConfig `mapstructure:"smtp" validate:"required"`
}

type SMTPConfig struct {
	Host       string `mapstructure:"host" validate:"required"`
	Port       int    `mapstructure:"port" validate:"required"`
	Username   string `mapstructure:"username" validate:"required"`
	Password   string `mapstructure:"password" validate:"required"`
	From       string `mapstructure:"from" validate:"required"`
	Timeout    int    `mapstructure:"timeout" validate:"required"`    // seconds
	Encryption string `mapstructure:"encryption" validate:"required"` // none, tls, starttls
}

type LoggingConfig struct {
	Level      string `mapstructure:"level" validate:"required"`
	Format     string `mapstructure:"format" validate:"required"`
	FilePath   string `mapstructure:"file_path" validate:"required"`
	MaxSizeMB  int    `mapstructure:"max_size_mb" validate:"required"`
	MaxBackups int    `mapstructure:"max_backups" validate:"required"`
	MaxAgeDays int    `mapstructure:"max_age_days" validate:"required"`
	Compress   bool   `mapstructure:"compress" validate:"required"`
}

type StorageConfig struct {
	Type      string `mapstructure:"type" validate:"required"`
	Endpoint  string `mapstructure:"endpoint" validate:"required"`
	Bucket    string `mapstructure:"bucket" validate:"required"`
	AccessKey string `mapstructure:"access_key" validate:"required"`
	SecretKey string `mapstructure:"secret_key" validate:"required"`
	UseSSL    bool   `mapstructure:"use_ssl" validate:"required"`
}

type AntiReplayConfig struct {
	Enabled            bool `mapstructure:"enabled" validate:"required"`
	TimestampTolerance int  `mapstructure:"timestamp_tolerance" validate:"required"` // minutes
	CacheTTL           int  `mapstructure:"cache_ttl" validate:"required"`           // seconds
}

type CORSConfig struct {
	AllowedOrigins   []string `mapstructure:"allowed_origins" validate:"required"`
	AllowedMethods   []string `mapstructure:"allowed_methods" validate:"required"`
	AllowedHeaders   []string `mapstructure:"allowed_headers" validate:"required"`
	AllowCredentials bool     `mapstructure:"allow_credentials" validate:"required"`
	MaxAge           int      `mapstructure:"max_age" validate:"required"`
}

type RateLimitConfig struct {
	Enabled                  bool `mapstructure:"enabled" validate:"required"`
	DefaultRequestsPerMinute int  `mapstructure:"default_requests_per_minute" validate:"required"`
	Burst                    int  `mapstructure:"burst" validate:"required"`
}

type CoreConfig struct {
	Server     ServerConfig     `mapstructure:"server" validate:"required"`
	CORS       CORSConfig       `mapstructure:"cors" validate:"required"`
	Logging    LoggingConfig    `mapstructure:"logging" validate:"required"`
	Database   DatabaseConfig   `mapstructure:"database" validate:"required"`
	Auth       AuthConfig       `mapstructure:"auth" validate:"required"`
	Cache      CacheConfig      `mapstructure:"cache" validate:"required"`
	Storage    StorageConfig    `mapstructure:"storage" validate:"required"`
	Email      EmailConfig      `mapstructure:"email" validate:"required"`
	RateLimit  RateLimitConfig  `mapstructure:"rate_limit" validate:"required"`
	AntiReplay AntiReplayConfig `mapstructure:"antireplay" validate:"required"`
}

func NewConfig[T any](configFilePath, configDir string) (*T, error) {
	return loadConfig[T](configDir, configFilePath)
}

func loadConfig[T any](configDir, configFilePath string) (*T, error) {
	loadDotEnv()
	setupViper(configDir, configFilePath)
	readConfigFile()
	bindEnvVariables()

	var cfg T
	if err := viper.UnmarshalExact(&cfg); err != nil {
		return nil, err
	}

	// Validate the struct fields
	validate := validator.New()
	if err := validate.Struct(&cfg); err != nil {
		return nil, fmt.Errorf("config validation failed: %w\n", err)
	}

	return &cfg, nil

}

func loadDotEnv() {
	_ = godotenv.Load()
}

func setupViper(configDir, configFilePath string) {
	if configFilePath != "" {
		viper.SetConfigFile(configFilePath)
	} else {
		viper.SetConfigName("config")
		viper.SetConfigType("yaml")
		viper.AddConfigPath(configDir)
		viper.AddConfigPath(".")
	}
}

func readConfigFile() {
	_ = viper.ReadInConfig()
}

func bindEnvVariables() {
	viper.SetEnvPrefix("APP")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()
}
