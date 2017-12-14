package multyback

import (
	"io/ioutil"
	"os"

	"github.com/go-yaml/yaml"
)

// TODO: add yaml tags to structures

// JWTConf is struct with JWT parameters
type JWTConf struct {
	Secret      string
	ElapsedDays int
}

// Configuration is a struct with all service options
type Configuration struct {
	Name             string
	Address          string
	DataStore        map[string]interface{}
	JWTConfig        *JWTConf
	DataStoreAddress string
}

// GetConfig initializes configuration for multy backend
func GetConfig(confiFile string) (*Configuration, error) {
	fd, err := os.Open(confiFile)
	if err != nil {
		return nil, err
	}
	rawData, err := ioutil.ReadAll(fd)
	if err != nil {
		return nil, err
	}

	conf := &Configuration{}
	err = yaml.Unmarshal([]byte(rawData), conf)
	if err != nil {
		return nil, err
	}

	return conf, nil
}
