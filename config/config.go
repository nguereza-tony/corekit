package config

import (
	"fmt"
	"strings"

	"github.com/go-playground/locales/en"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	en_translations "github.com/go-playground/validator/v10/translations/en"
	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

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

	enLocale := en.New()
	uni := ut.New(enLocale, enLocale)
	trans, _ := uni.GetTranslator("en")

	// Validate the struct fields
	validate := validator.New()
	en_translations.RegisterDefaultTranslations(validate, trans)
	if err := validate.Struct(&cfg); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			for _, e := range validationErrors {
				fmt.Printf("%s: %s\n", e.StructNamespace(), e.Translate(trans))
			}
		}
		return nil, fmt.Errorf("config items validation failed")
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
