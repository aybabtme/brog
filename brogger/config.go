package brogger

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"runtime"
	"strings"
)

const (
	// ConfigFilename where to find the Brog config file.
	ConfigFilename = "./brog_config.json"
	// JSONIndentCount how many spaces to indent the config file
	JSONIndentCount = 3
)

// Defaults for Brog's configuration.
var (
	DefaultPortNumber   = 3000
	DefaultHostname     = "localhost"
	DefaultMaxCPUs      = runtime.NumCPU()
	DefaultTemplatePath = "templates/"
	DefaultPostPath     = "posts/"
	DefaultLogFilename  = "brog.log"
)

type Config struct {
	PortNumber   int    `json:"portNumber"`
	Hostname     string `json:"hostName"`
	MaxCPUs      int    `json:"maxCpus"`
	TemplatePath string `json:"templatePath"`
	PostPath     string `json:"postPath"`
	LogFilename  string `json:"logFilename"`
}

func (c *Config) selfValidate() error {
	if c.PortNumber < 1 || c.PortNumber > 1<<16 {
		return fmt.Errorf("invalid port range (%d)", c.PortNumber)
	}

	if c.Hostname == "" {
		return fmt.Errorf("invalid hostname (%s)", c.Hostname)
	}

	if c.MaxCPUs < 0 {
		return fmt.Errorf("invalid CPU count (%d)", c.MaxCPUs)
	}

	c.PostPath = path.Clean(c.PostPath)
	c.TemplatePath = path.Clean(c.TemplatePath)
	c.LogFilename = path.Clean(c.LogFilename)

	return nil
}

func loadConfig() (*Config, error) {
	if fileExists(ConfigFilename) {
		return loadFromFile()
	}

	c := &Config{
		PortNumber:   DefaultPortNumber,
		Hostname:     DefaultHostname,
		MaxCPUs:      DefaultMaxCPUs,
		TemplatePath: path.Clean(DefaultTemplatePath),
		PostPath:     path.Clean(DefaultPostPath),
		LogFilename:  path.Clean(DefaultLogFilename),
	}

	err := persistToFile(ConfigFilename, c)

	return c, err
}

func fileExists(filename string) bool {
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return false
	}
	return true
}

func loadFromFile() (*Config, error) {
	configRd, err := os.Open(ConfigFilename)
	if err != nil {
		return nil, fmt.Errorf("opening config file '%s', %v", ConfigFilename, err)
	}
	defer configRd.Close()

	jsonDec := json.NewDecoder(configRd)
	config := Config{}

	err = jsonDec.Decode(&config)
	if err != nil {
		return nil, fmt.Errorf("decoding config file, %v", err)
	}

	err = config.selfValidate()
	if err != nil {
		return nil, fmt.Errorf("validating config settings, %v", err)
	}

	return &config, nil
}

func persistToFile(filename string, config *Config) error {

	jsonData, err := json.MarshalIndent(config, "", strings.Repeat(" ", JSONIndentCount))
	if err != nil {
		return fmt.Errorf("marshalling config data, %v", err)
	}

	confFile, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("creating file '%s' for configuration, %v", filename, err)
	}
	defer confFile.Close()

	_, err = confFile.WriteString(string(jsonData))
	if err != nil {
		return fmt.Errorf("writing configuration to file '%s', %v", filename, err)
	}
	return nil
}
