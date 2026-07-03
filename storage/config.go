package storage

type StorageConfig struct {
	Type      string
	Endpoint  string
	Bucket    string
	AccessKey string
	SecretKey string
	UseSSL    bool
}
