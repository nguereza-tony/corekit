package logger

type LoggingConfig struct {
	Level      string
	Format     string
	FilePath   string
	MaxSizeMB  int
	MaxBackups int
	MaxAgeDays int
	Compress   bool
}
