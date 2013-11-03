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
	// JSONIndentCount how many spaces to indent the config file
	JSONIndentCount = 3
)

// Defaults for Brog's configuration.
var (
	DefaultPortNumber   = 3000
	DefaultHostname     = mustHave(os.Hostname())
	DefaultMaxCPUs      = runtime.NumCPU()
	DefaultTemplatePath = "templates/"
	DefaultPostPath     = "posts/"
	DefaultAssetPath    = "assets/"
	DefaultLogFilename  = "brog.log"
)

func mustHave(value string, err error) string {
	if err != nil {
		panic(err)
	}
	return value
}

// Config contains all the settings that a Brog uses to watch and create
// and serve posts, log events and execute in general.
type Config struct {
	PortNumber   int    `json:"portNumber"`
	Hostname     string `json:"hostName"`
	MaxCPUs      int    `json:"maxCpus"`
	TemplatePath string `json:"templatePath"`
	PostPath     string `json:"postPath"`
	AssetPath    string `json:"assetPath"`
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
	c.AssetPath = path.Clean(c.AssetPath)
	c.PostPath = path.Clean(c.PostPath)
	c.TemplatePath = path.Clean(c.TemplatePath)
	c.LogFilename = path.Clean(c.LogFilename)

	return nil
}

func loadConfig() (*Config, error) {
	if !fileExists(ConfigFilename) {
		return nil, fmt.Errorf("There is no Brog config file here.")

	}
	return loadFromFile()
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

	if err := configRd.Close(); err != nil {
		return &config, fmt.Errorf("closing config file, %v", err)
	}

	return &config, nil
}

func newDefaultConfig() *Config {
	return &Config{
		PortNumber:   DefaultPortNumber,
		Hostname:     DefaultHostname,
		MaxCPUs:      DefaultMaxCPUs,
		TemplatePath: path.Clean(DefaultTemplatePath),
		PostPath:     path.Clean(DefaultPostPath),
		AssetPath:    path.Clean(DefaultAssetPath),
		LogFilename:  path.Clean(DefaultLogFilename),
	}
}

func (config *Config) persistToFile(filename string) error {

	jsonData, err := json.MarshalIndent(config, "", strings.Repeat(" ", JSONIndentCount))
	if err != nil {
		return fmt.Errorf("marshalling config data, %v", err)
	}

	confFile, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("creating file '%s' for configuration, %v", filename, err)
	}

	_, err = confFile.WriteString(string(jsonData))
	if err != nil {
		return fmt.Errorf("writing configuration to file '%s', %v", filename, err)
	}

	if err := confFile.Close(); err != nil {
		return fmt.Errorf("closing config file '%s', %v", filename, err)
	}

	return nil
}
