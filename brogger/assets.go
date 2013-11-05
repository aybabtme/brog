package brogger

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
)

const (
	// ConfigFilename where to find the Brog config file.
	ConfigFilename = "brog_config.json"
	jsPath         = "js/"
	cssPath        = "css/"
)

// Base templates
var allAssets = map[string]packed{
	// Base CSS
	"brog.css":   {"brog.css", path.Join(DefaultAssetPath, cssPath), baseAssetsCssBrogCss},
	"github.css": {"github.css", path.Join(DefaultAssetPath, cssPath), baseAssetsCssGithubCss},

	// Base JS,
	"brog.js":          {"brog.js", path.Join(DefaultAssetPath, jsPath), baseAssetsJsBrogJs},
	"highlight.min.js": {"highlight.min.js", path.Join(DefaultAssetPath, jsPath), baseAssetsJsHighlightMinJs},

	// Base posts
	"sample.md": {"sample.md", DefaultPostPath, basePostsSampleMd},
	"blank.md":  {"blank.md", DefaultPostPath, basePostsBlankMd},

	// Base files
	".gitignore": {".gitignore", "./", baseGitignore},
	"README.md":  {"README.md", "./", baseREADMEMd},
}

var allTemplates = map[string]packed{
	// HTML templates
	appTmplName:        {appTmplName, DefaultTemplatePath, baseTemplatesApplicationGohtml},
	indexTmplName:      {indexTmplName, DefaultTemplatePath, baseTemplatesIndexGohtml},
	postTmplName:       {postTmplName, DefaultTemplatePath, baseTemplatesPostGohtml},
	langSelectTmplName: {langSelectTmplName, DefaultTemplatePath, baseTemplatesLangselectGohtml},
	styleTmplName:      {styleTmplName, DefaultTemplatePath, baseTemplatesStyleGohtml},
	jsTmplName:         {jsTmplName, DefaultTemplatePath, baseTemplatesJavascriptGohtml},
	headerTmplName:     {headerTmplName, DefaultTemplatePath, baseTemplatesHeaderGohtml},
	footerTmplName:     {footerTmplName, DefaultTemplatePath, baseTemplatesFooterGohtml},
}

type packed struct {
	filename    string
	destination string
	data        []byte
}

func (p *packed) replicate() error {
	fullpath := path.Join(p.destination, p.filename)
	if fileExists(fullpath) {
		return fmt.Errorf("file '%s' already exists, will not overwrite", fullpath)
	}
	return p.rewriteFile(fullpath)
}

func (p *packed) replicateInDir(dirname string) error {
	fullpath := path.Join(dirname, p.filename)
	if fileExists(fullpath) {
		return fmt.Errorf("file '%s' already exists, will not overwrite", fullpath)
	}
	return p.rewriteFile(fullpath)
}

func (p *packed) rewriteInDir(dirname string) error {
	fullpath := path.Join(dirname, p.filename)
	return p.rewriteFile(fullpath)
}

func (p *packed) rewriteFile(fullpath string) error {
	if err := os.MkdirAll(p.destination, 0740); err != nil {
		return fmt.Errorf("creating directory at '%s' failed, %v", p.destination, err)
	}
	return ioutil.WriteFile(fullpath, p.data, 0640)
}

// CopyBrogBinaries writes the in-memory brog assets to the current working
// directory, effectively creating a brog structure that `brog server` can use
// to serve content.
func CopyBrogBinaries() []error {

	config := newDefaultConfig()
	errs := []error{}

	err := config.persistToFile(ConfigFilename)
	if err != nil {
		fmt.Errorf("persisting config file, %v", err)
	}

	for _, asset := range allAssets {
		err := asset.replicate()
		if err != nil {
			errs = append(errs, err)
		}
	}

	for _, asset := range allTemplates {
		err := asset.replicate()
		if err != nil {
			errs = append(errs, err)
		}
	}

	return errs
}

// CopyBlankToFilename creates a blank post at the given filename, under the asset
// path specified by conf
func CopyBlankToFilename(conf *Config, filename string) error {
	if filename == "" {
		return fmt.Errorf("no filename specified")
	}
	fullpath := path.Clean(conf.PostPath) + string(os.PathSeparator) + filename + conf.PostFileExt
	return ioutil.WriteFile(fullpath, basePostsBlankMd, 0640)
}
