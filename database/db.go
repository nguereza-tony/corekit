package database

import (
	"fmt"
	"io"
	"log"
	"os"
	"sync"
	"time"

	"github.com/natefinch/lumberjack"
	"github.com/nguereza-tony/corekit/config"
	"github.com/nguereza-tony/corekit/logger"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/driver/sqlserver"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

var (
	dbInstance *gorm.DB
	dbOnce     sync.Once
)

// Connect initializes and returns the database connection (singleton)
func NewDatabase(
	dbConfig *config.DatabaseConfig,
	loggerConfig *config.LoggingConfig,
	logger *logger.Logger,
) (*gorm.DB, error) {
	var err error
	dbOnce.Do(func() {
		dsn := buildDSN(dbConfig)

		var driver gorm.Dialector
		switch dbConfig.Driver {
		case "mysql":
			driver = mysql.Open(dsn)
		case "postgres":
			driver = postgres.Open(dsn)
		case "sqlite":
			driver = sqlite.Open(dsn)
		default:
			driver = sqlserver.Open(dsn)
		}
		db, connErr := gorm.Open(driver, &gorm.Config{
			Logger:         getLogger(dbConfig.Logging),
			PrepareStmt:    true,
			TranslateError: true,
		})
		if connErr != nil {
			err = fmt.Errorf("failed to connect to database: %w", connErr)
			return
		}

		// Get underlying sql.DB to configure connection pool
		sqlDB, sqlErr := db.DB()
		if sqlErr != nil {
			err = fmt.Errorf("failed to get sql.DB: %w", sqlErr)
			return
		}

		// Configure connection pool
		sqlDB.SetMaxOpenConns(dbConfig.MaxOpenConns)
		sqlDB.SetMaxIdleConns(dbConfig.MaxIdleConns)
		sqlDB.SetConnMaxLifetime(time.Duration(dbConfig.ConnMaxLifetime) * time.Second)

		dbInstance = db
		logger.Infof("Database connection established (pool: max_open=%d, max_idle=%d, max_lifetime=%ds)",
			dbConfig.MaxOpenConns, dbConfig.MaxIdleConns, dbConfig.ConnMaxLifetime)
	})
	return dbInstance, err
}

// Get returns the existing database connection (panics if not initialized)
func GetDbInstance() *gorm.DB {
	if dbInstance == nil {
		panic("database not initialized. Call Connect first")
	}
	return dbInstance
}

// buildDSN constructs the connection string
func buildDSN(cfg *config.DatabaseConfig) string {
	if cfg.Driver == "mysql" {
		return fmt.Sprintf(
			"%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.DBName,
		)
	}
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName, cfg.SSLMode,
	)
}

// getGORMLogLevel maps our log level to GORM's log level
func getGORMLogLevel(level string) gormlogger.LogLevel {
	switch level {
	case "debug":
		return gormlogger.Info
	case "info":
		return gormlogger.Warn
	default:
		return gormlogger.Error
	}
}

func getLogger(cfg config.DatabaseLoggingConfig) gormlogger.Interface {
	var writers []io.Writer
	if cfg.Console {
		writers = append(writers, os.Stdout)
	}

	if cfg.FilePath != "" {
		writers = append(writers,
			&lumberjack.Logger{
				Filename:   cfg.FilePath,
				MaxSize:    cfg.MaxSizeMB,
				MaxBackups: cfg.MaxBackups,
				MaxAge:     cfg.MaxAgeDays,
				Compress:   cfg.Compress,
			})
	}

	// Create a new standard logger pointing to our MultiWriter
	// This ensures that the output is properly formatted and safe for concurrent use
	newLog := log.New(io.MultiWriter(writers...), "\r\n", log.LstdFlags)

	// Wrap it in a GORM Logger
	logger := gormlogger.New(
		newLog,
		gormlogger.Config{
			SlowThreshold:             time.Duration(cfg.SlowSqlThreshold) * time.Second,
			LogLevel:                  getGORMLogLevel(cfg.Level),
			IgnoreRecordNotFoundError: true,
			ParameterizedQueries:      true,
			Colorful:                  false,
		},
	)

	return logger
}
