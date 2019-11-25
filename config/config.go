package config

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

type Config struct {
	Server struct {
		Addr string `yaml:"addr"`
	}
	Database struct {
		DSN         string `yaml:"dsn"`
		MaxIdleConn int    `yaml:"max_idle_conn"`
	}
	Minio struct {
		Endpoint        string `yaml:"endpoint"`
		AccessKeyID     string `yaml:"accessKeyID"`
		SecretAccessKey string `yaml:"secretAccessKey"`
		UseSSL          bool   `yaml:"useSSL"`
		Bucket          string `yaml:"bucket"`
	}
	Nsq struct {
		Endpoint string `yaml:"endpoint"`
		Topic    string `yaml:"topic"`
		Channel  string `yaml:"channel"`
	}
}

var config *Config

func Load(path string) error {
	res, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}

	return yaml.Unmarshal(res, &config)
}

func Get() *Config {
	return config
}
