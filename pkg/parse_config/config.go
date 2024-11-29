package parse_config

import (
	"github.com/go-playground/validator/v10"
	"gopkg.in/yaml.v3"
	"log"
	"os"
)

type Config struct {
	Providers []Provider `yaml:"providers"`
	Global
}

type Global struct {
	CertStoragePath string `yaml:"certStoragePath"`
}

type Provider struct {
	ProviderName   string   `yaml:"providerName" validate:"required"`
	AccessKey      string   `yaml:"accessKey" validate:"required"`
	SecretKey      string   `yaml:"secretKey"`
	Domains        []string `yaml:"domains" validate:"required"`
	Enable         bool     `yaml:"enable" validate:"required"`
	RenewBeforeDay int      `yaml:"renewBeforeDay"`
	Email          string   `yaml:"email" validate:"required"`
	Hook           string   `yaml:"hook"`
	SaveSSLName    string   `yaml:"saveSSLName"`
}

func ParseConfig() (Config, error) {
	absPath, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	configPath := absPath + "/" + "config.yml"
	data, err := os.ReadFile(configPath)
	if err != nil {
		log.Fatalf("Error Reading YAML File: %v.", err)
		return Config{}, err
	}
	log.Println("read config successes, from: ", configPath)

	// Parse the YAML file
	var config Config
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		log.Fatalf("Error parsing YAML file: %v", err)
		return Config{}, err
	}
	validate := validator.New()
	err = validate.Struct(config)
	if err != nil {
		return Config{}, err
	}
	return config, nil
}
