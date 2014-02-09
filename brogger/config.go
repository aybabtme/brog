package brogger

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
)

const (
	Version = "v0.0.2"
	// JSONIndentCount how many spaces to indent the config file
	JSONIndentCount = 3
)

// Defaults for Brog's configuration.
var (
	DefaultProdPort         = "80"
	DefaultDevelPort        = "3000"
	DefaultHostname         = "localhost"
	DefaultMaxCPUs          = runtime.NumCPU()
	DefaultTemplatePath     = "templates" + string(os.PathSeparator)
	DefaultPostPath         = "posts" + string(os.PathSeparator)
	DefaultPagePath         = "pages" + string(os.PathSeparator)
	DefaultAssetPath        = "assets" + string(os.PathSeparator)
	DefaultPostFileExt      = ".md"
	DefaultLogFilename      = "brog.log"
	DefaultLogVerbosity     = "watch"
	DefaultConsoleVerbosity = "watch"
	DefaultRewriteInvalid   = true  // True so that brog has stable default
	DefaultRewriteMissing   = true  // True so that brog has stable default
	DefaultMultilingual     = false // False because blogs are usually unilingual
	DefaultLanguages        = []string{"en"}
)

// Config contains all the settings that a Brog uses to watch and create
// and serve posts, log events and execute in general.
type Config struct {
	ProdPort         string   `json:"prodPort"`
	DevelPort        string   `json:"develPort"`
	Hostname         string   `json:"hostName"`
	MaxCPUs          int      `json:"maxCpus"`
	TemplatePath     string   `json:"templatePath"`
	PostPath         string   `json:"postPath"`
	PagePath         string   `json:"pagePath"`
	AssetPath        string   `json:"assetPath"`
	PostFileExt      string   `json:"postFileExtension"`
	LogFilename      string   `json:"logFilename"`
	LogFileVerbosity string   `json:"logFileVerbosity"`
	ConsoleVerbosity string   `json:"consoleVerbosity"`
	RewriteInvalid   bool     `json:"rewriteInvalid"`
	RewriteMissing   bool     `json:"rewriteMissing"`
	Multilingual     bool     `json:"multilingual"`
	Languages        []string `json:"languages"`
}

func newDefaultConfig() *Config {
	return &Config{
		ProdPort:         DefaultProdPort,
		DevelPort:        DefaultDevelPort,
		Hostname:         DefaultHostname,
		MaxCPUs:          DefaultMaxCPUs,
		TemplatePath:     filepath.Clean(DefaultTemplatePath),
		PostPath:         filepath.Clean(DefaultPostPath),
		PagePath:         filepath.Clean(DefaultPagePath),
		AssetPath:        filepath.Clean(DefaultAssetPath),
		PostFileExt:      DefaultPostFileExt,
		LogFilename:      filepath.Clean(DefaultLogFilename),
		LogFileVerbosity: DefaultLogVerbosity,
		ConsoleVerbosity: DefaultConsoleVerbosity,
		RewriteInvalid:   DefaultRewriteInvalid,
		RewriteMissing:   DefaultRewriteMissing,
		Multilingual:     DefaultMultilingual,
		Languages:        DefaultLanguages,
	}
}

func (cfg *Config) selfValidate() error {
	portnum, err := strconv.ParseInt(cfg.ProdPort, 10, 0)
	if err == nil && (portnum < 1 || portnum > 1<<16) {
		return fmt.Errorf("invalid port range (%d)", portnum)
	}
	portnum, err = strconv.ParseInt(cfg.DevelPort, 10, 64)
	if err == nil && (portnum < 1 || portnum > 1<<16) {
		return fmt.Errorf("invalid port range (%d)", portnum)
	}

	if cfg.MaxCPUs < 0 {
		return fmt.Errorf("invalid CPU count (%d)", cfg.MaxCPUs)
	}

	if cfg.PostFileExt == "" {
		return fmt.Errorf("invalid Post file extension (%s)", cfg.PostFileExt)
	}
	cfg.AssetPath = filepath.Clean(cfg.AssetPath)
	cfg.PostPath = filepath.Clean(cfg.PostPath)
	cfg.TemplatePath = filepath.Clean(cfg.TemplatePath)
	cfg.LogFilename = filepath.Clean(cfg.LogFilename)

	return nil
}

func loadConfig() (*Config, error) {
	if !fileExists(ConfigFilename) {
		return nil, fmt.Errorf("there is no brog config file named '%s' here", ConfigFilename)
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

func (cfg *Config) persistToFile(filename string) error {
	jsonData, err := json.MarshalIndent(cfg, "", strings.Repeat(" ", JSONIndentCount))
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

	/* Add newline to the end of file */
	_, err = confFile.WriteString("\n")
	if err != nil {
		return fmt.Errorf("writing newline to configuration file '%s', %v", filename, err)
	}

	if err := confFile.Close(); err != nil {
		return fmt.Errorf("closing config file '%s', %v", filename, err)
	}

	return nil
}
