package main

import (
	"errors"
	"io/ioutil"
	"log"
	"os"

	"github.com/gonuts/yaml"
)

type Config struct {
	Ftp_host              string
	Ftp_path              string
	Ftp_user              string
	Ftp_pass              string
	Ftp_proxy             string
	Scp_host              string
	Scp_path              string
	Scp_user              string
	Scp_port              string
	Archive_method        string
	Keep_failed_container bool
}

func GetYAML() ([]byte, error) {
	switch len(os.Args) {
	case 1:
		return ioutil.ReadAll(os.Stdin)
	case 2:
		return ioutil.ReadFile(os.Args[1])
	default:
		return nil, errors.New("Too many arguments.")
	}
}

func GetConfig() Config {
	yml, err := GetYAML()
	if err != nil {
		log.Fatal(err.Error())
	}

	var config Config

	err = yaml.Unmarshal(yml, &config)
	if err != nil {
		log.Fatal(err.Error())
	}

	return config
}
