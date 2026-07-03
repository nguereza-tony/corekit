package database

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
