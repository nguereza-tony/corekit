package config

import (
	"strings"
	"sync"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

type RedisConfig struct {
	Host         string `mapstructure:"host"`
	Port         int    `mapstructure:"port"`
	Password     string `mapstructure:"password"`
	DB           int    `mapstructure:"db"`
	PoolSize     int    `mapstructure:"pool_size"`
	MinIdleConns int    `mapstructure:"min_idle_conns"`
	DialTimeout  int    `mapstructure:"dial_timeout"`
	ReadTimeout  int    `mapstructure:"read_timeout"`
	WriteTimeout int    `mapstructure:"write_timeout"`
	MaxRetries   int    `mapstructure:"max_retries"`
}

type ServerConfig struct {
	Host        string `mapstructure:"host"`
	Port        int    `mapstructure:"port"`
	Mode        string `mapstructure:"mode"`
	DocsEnabled bool   `mapstructure:"docs_enabled"`
}

type AuthConfig struct {
	JwtSecret     string `mapstructure:"jwt_secret"`
	JwtAccessTTL  int    `mapstructure:"jwt_expire_time"`         // seconds
	JwtRefreshTTL int    `mapstructure:"jwt_refresh_expire_time"` // seconds
}

type MemoryCacheConfig struct {
	CleanupInterval int `mapstructure:"cleanup_interval"`
	MaxItems        int `mapstructure:"max_items"`
}

type CacheConfig struct {
	Type       string            `mapstructure:"type"`
	Prefix     string            `mapstructure:"prefix"`
	DefaultTTL int               `mapstructure:"default_ttl"`
	Redis      RedisConfig       `mapstructure:"redis"`
	Memory     MemoryCacheConfig `mapstructure:"memory"`
}

type DatabaseConfig struct {
	Driver          string `mapstructure:"driver"` // mysql, postgres, sqlite, sqlserver
	Host            string `mapstructure:"host"`
	Port            int    `mapstructure:"port"`
	User            string `mapstructure:"user"`
	Password        string `mapstructure:"password"`
	DBName          string `mapstructure:"dbname"`
	SSLMode         string `mapstructure:"sslmode"`
	MaxOpenConns    int    `mapstructure:"max_open_conns"`
	MaxIdleConns    int    `mapstructure:"max_idle_conns"`
	ConnMaxLifetime int    `mapstructure:"conn_max_lifetime"` // seconds
}

type EmailConfig struct {
	Provider string     `mapstructure:"provider"`
	SMTP     SMTPConfig `mapstructure:"smtp"`
}

type SMTPConfig struct {
	Host       string `mapstructure:"host"`
	Port       int    `mapstructure:"port"`
	Username   string `mapstructure:"username"`
	Password   string `mapstructure:"password"`
	From       string `mapstructure:"from"`
	Timeout    int    `mapstructure:"timeout"`    // seconds
	Encryption string `mapstructure:"encryption"` // none, tls, starttls
}

type LoggingConfig struct {
	Level      string `mapstructure:"level"`
	Format     string `mapstructure:"format"`
	FilePath   string `mapstructure:"file_path"`
	MaxSizeMB  int    `mapstructure:"max_size_mb"`
	MaxBackups int    `mapstructure:"max_backups"`
	MaxAgeDays int    `mapstructure:"max_age_days"`
	Compress   bool   `mapstructure:"compress"`
}

type StorageConfig struct {
	Type      string `mapstructure:"type"`
	Endpoint  string `mapstructure:"endpoint"`
	Bucket    string `mapstructure:"bucket"`
	AccessKey string `mapstructure:"access_key"`
	SecretKey string `mapstructure:"secret_key"`
	UseSSL    bool   `mapstructure:"use_ssl"`
}

type AntiReplayConfig struct {
	Enabled            bool `mapstructure:"enabled"`
	TimestampTolerance int  `mapstructure:"timestamp_tolerance"` // minutes
	CacheTTL           int  `mapstructure:"cache_ttl"`           // seconds
}

type CORSConfig struct {
	AllowedOrigins   []string `mapstructure:"allowed_origins"`
	AllowedMethods   []string `mapstructure:"allowed_methods"`
	AllowedHeaders   []string `mapstructure:"allowed_headers"`
	AllowCredentials bool     `mapstructure:"allow_credentials"`
	MaxAge           int      `mapstructure:"max_age"`
}

type RateLimitConfig struct {
	Enabled                  bool `mapstructure:"enabled"`
	DefaultRequestsPerMinute int  `mapstructure:"default_requests_per_minute"`
	Burst                    int  `mapstructure:"burst"`
}

type Config struct {
	Server     ServerConfig     `mapstructure:"server"`
	CORS       CORSConfig       `mapstructure:"cors"`
	Logging    LoggingConfig    `mapstructure:"database"`
	Database   DatabaseConfig   `mapstructure:"database"`
	Auth       AuthConfig       `mapstructure:"auth"`
	Cache      CacheConfig      `mapstructure:"cache"`
	Storage    StorageConfig    `mapstructure:"storage"`
	Email      EmailConfig      `mapstructure:"email"`
	RateLimit  RateLimitConfig  `mapstructure:"rate_limit"`
	AntiReplay AntiReplayConfig `mapstructure:"antireplay"`
}

var (
	configInstance *Config
	configOnce     sync.Once
)

func NewConfig(configFilePath, configDir string) *Config {
	configOnce.Do(func() {
		configInstance = loadConfig(configDir, configFilePath)
	})
	return configInstance
}

func loadConfig(configDir, configFilePath string) *Config {
	loadDotEnv()
	setupViper(configDir, configFilePath)
	readConfigFile()
	bindEnvVariables()

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		panic(err)
	}

	return &cfg

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
