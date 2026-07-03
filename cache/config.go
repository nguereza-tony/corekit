package cache

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
