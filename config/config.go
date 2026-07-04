package config

type RedisConfig struct {
	Host         string
	Port         int
	Password     string
	DB           int
	PoolSize     int
	MinIdleConns int
	DialTimeout  int
	ReadTimeout  int
	WriteTimeout int
	MaxRetries   int
}

type AuthConfig struct {
	JwtSecret     string
	JwtAccessTTL  int // seconds
	JwtRefreshTTL int // seconds
}

type MemoryCacheConfig struct {
	CleanupInterval int
	MaxItems        int
}

type CacheConfig struct {
	Type       string
	Prefix     string
	DefaultTTL int
	Redis      RedisConfig
	Memory     MemoryCacheConfig
}

type DatabaseConfig struct {
	Driver          string // mysql, postgres, sqlite, sqlserver
	Host            string
	Port            int
	User            string
	Password        string
	DBName          string
	SSLMode         string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime int // seconds
}

type EmailConfig struct {
	Provider string
	SMTP     SMTPConfig
}

type SMTPConfig struct {
	Host       string
	Port       int
	Username   string
	Password   string
	From       string
	Timeout    int    // seconds
	Encryption string // none, tls, starttls
}

type LoggingConfig struct {
	Level      string
	Format     string
	FilePath   string
	MaxSizeMB  int
	MaxBackups int
	MaxAgeDays int
	Compress   bool
}

type StorageConfig struct {
	Type      string
	Endpoint  string
	Bucket    string
	AccessKey string
	SecretKey string
	UseSSL    bool
}

type AntiReplayConfig struct {
	Enabled            bool
	TimestampTolerance int // minutes
	CacheTTL           int // seconds
}

type CORSConfig struct {
	AllowedOrigins   []string
	AllowedMethods   []string
	AllowedHeaders   []string
	AllowCredentials bool
	MaxAge           int
}
