package logger

import (
	"io"
	"os"
	"sync"

	"github.com/natefinch/lumberjack"
	"github.com/sirupsen/logrus"
)

type Logger struct {
	entry *logrus.Entry
}

var (
	loggerInstance *Logger
	loggerOnce     sync.Once
)

func NewLogger(cfg LoggingConfig) *Logger {
	loggerOnce.Do(func() {
		log := logrus.New()

		// Set level
		level, err := logrus.ParseLevel(cfg.Level)
		if err != nil {
			level = logrus.DebugLevel
		}
		log.SetLevel(level)

		// Set formatter
		switch cfg.Format {
		case "json":
			log.SetFormatter(&logrus.JSONFormatter{})
		case "text":
			log.SetFormatter(&logrus.TextFormatter{
				FullTimestamp: true,
				ForceColors:   true,
				DisableQuote:  true,
			})
		default:
			log.SetFormatter(&logrus.JSONFormatter{})
		}

		// Configure output (file + console)
		var writers []io.Writer
		writers = append(writers, os.Stdout)

		if cfg.FilePath != "" {
			writers = append(writers, &lumberjack.Logger{
				Filename:   cfg.FilePath,
				MaxSize:    cfg.MaxSizeMB,
				MaxBackups: cfg.MaxBackups,
				MaxAge:     cfg.MaxAgeDays,
				Compress:   cfg.Compress,
			})
		}

		log.SetOutput(io.MultiWriter(writers...))

		loggerInstance = &Logger{
			entry: logrus.NewEntry(log),
		}
	})
	return loggerInstance
}

// Methods
func (l *Logger) Info(args ...interface{}) {
	l.entry.Info(args...)
}

func (l *Logger) Infof(format string, args ...interface{}) {
	l.entry.Infof(format, args...)
}

func (l *Logger) Debug(args ...interface{}) {
	l.entry.Debug(args...)
}

func (l *Logger) Debugf(format string, args ...interface{}) {
	l.entry.Debugf(format, args...)
}

func (l *Logger) Warn(args ...interface{}) {
	l.entry.Warn(args...)
}

func (l *Logger) Warnf(format string, args ...interface{}) {
	l.entry.Warnf(format, args...)
}

func (l *Logger) Error(args ...interface{}) {
	l.entry.Error(args...)
}

func (l *Logger) Errorf(format string, args ...interface{}) {
	l.entry.Errorf(format, args...)
}

func (l *Logger) Fatal(args ...interface{}) {
	l.entry.Fatal(args...)
}

func (l *Logger) Fatalf(format string, args ...interface{}) {
	l.entry.Fatalf(format, args...)
}
