package config

import (
	"github.com/go-emd/emd/log"
	"encoding/json"
	"io/ioutil"
)

type Connection struct {
	Type   string
	Worker string
	Alias  string
	Buffer string
}

type WorkConfig struct {
	Name        string
	Connections []Connection
}

type NodeConfig struct {
	Hostname string
	Workers  []WorkConfig
}

type Config struct {
	Nfs      bool
	GUI_port string
	Nodes    []NodeConfig
}

func Process(path string, config *Config) {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		log.ERROR.Println("Unable to read config file.")
	}

	err = json.Unmarshal(b, &config)
	if err != nil {
		log.ERROR.Println("Unable to parse json config file.")
	}
}
