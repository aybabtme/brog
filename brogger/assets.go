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

type packed struct {
	filename string
	data     []byte
}

func (p *packed) replicateInDir(dirpath string) error {
	fullpath := path.Clean(dirpath) + string(os.PathSeparator) + p.filename
	if fileExists(fullpath) {
		return fmt.Errorf("file '%s' already exists, will not overwrite", fullpath)
	}
	return ioutil.WriteFile(fullpath, p.data, 0640)
}

func (p *packed) rewriteFileOnDir(dirpath string) error {
	fullpath := path.Clean(dirpath) + string(os.PathSeparator) + p.filename
	return ioutil.WriteFile(fullpath, p.data, 0640)
}

// Base templates
var (
	appPaktTmpl        = packed{appTmplName, baseTemplatesApplicationGohtml}
	indexPaktTmpl      = packed{indexTmplName, baseTemplatesIndexGohtml}
	postPaktTmpl       = packed{postTmplName, baseTemplatesPostGohtml}
	stylePaktTmpl      = packed{styleTmplName, baseTemplatesStyleGohtml}
	javascriptPaktTmpl = packed{jsTmplName, baseTemplatesJavascriptGohtml}
	headerPaktTmpl     = packed{headerTmplName, baseTemplatesHeaderGohtml}
	footerPaktTmpl     = packed{footerTmplName, baseTemplatesFooterGohtml}
)

// Base CSS
var brogCss = packed{cssPath + "brog.css", baseAssetsCssBrogCss}

// Base JS
var brogJs = packed{jsPath + "brog.js", baseAssetsJsBrogJs}

// Base posts
var (
	samplePost = packed{"sample.md", basePostsSampleMd}
	blankPost  = packed{"blank.md", basePostsBlankMd}
)

// CopyBrogBinaries writes the in-memory brog assets to the current working
// directory, effectively creating a brog structure that `brog server` can use
// to serve content.
func CopyBrogBinaries() error {

	config := newDefaultConfig()

	err := config.persistToFile(ConfigFilename)
	if err != nil {
		fmt.Errorf("persisting config file, %v", err)
	}

	// Posts
	if err := os.MkdirAll(DefaultPostPath, 0740); err != nil {
		return fmt.Errorf("creating directory at '%s' failed, %v", DefaultPostPath, err)
	}

	if err := samplePost.replicateInDir(DefaultPostPath); err != nil {
		return fmt.Errorf("replicating %s, %v", samplePost.filename, err)
	}

	// Assets
	cssDirPath := path.Join(DefaultAssetPath, cssPath)
	if err := os.MkdirAll(cssDirPath, 0740); err != nil {
		return fmt.Errorf("creating directory at '%s' failed, %v", cssDirPath, err)
	}
	jsDirPath := path.Join(DefaultAssetPath, jsPath)
	if err := os.MkdirAll(jsDirPath, 0740); err != nil {
		return fmt.Errorf("creating directory at '%s' failed, %v", jsDirPath, err)
	}
	if err := brogCss.replicateInDir(DefaultAssetPath); err != nil {
		return fmt.Errorf("replicating %s, %v", brogCss.filename, err)
	}
	if err := brogJs.replicateInDir(DefaultAssetPath); err != nil {
		return fmt.Errorf("replicating %s, %v", brogJs.filename, err)
	}

	// Templates
	if err := os.MkdirAll(DefaultTemplatePath, 0740); err != nil {
		return fmt.Errorf("creating directory at '%s' failed, %v", DefaultTemplatePath, err)
	}
	if err := appPaktTmpl.replicateInDir(DefaultTemplatePath); err != nil {
		return fmt.Errorf("replicating %s, %v", appPaktTmpl.filename, err)
	}
	if err := indexPaktTmpl.replicateInDir(DefaultTemplatePath); err != nil {
		return fmt.Errorf("replicating %s, %v", indexPaktTmpl.filename, err)
	}
	if err := postPaktTmpl.replicateInDir(DefaultTemplatePath); err != nil {
		return fmt.Errorf("replicating %s, %v", postPaktTmpl.filename, err)
	}
	if err := stylePaktTmpl.replicateInDir(DefaultTemplatePath); err != nil {
		return fmt.Errorf("replicating %s, %v", stylePaktTmpl.filename, err)
	}
	if err := javascriptPaktTmpl.replicateInDir(DefaultTemplatePath); err != nil {
		return fmt.Errorf("replicating %s, %v", javascriptPaktTmpl.filename, err)
	}
	if err := headerPaktTmpl.replicateInDir(DefaultTemplatePath); err != nil {
		return fmt.Errorf("replicating %s, %v", headerPaktTmpl.filename, err)
	}
	if err := footerPaktTmpl.replicateInDir(DefaultTemplatePath); err != nil {
		return fmt.Errorf("replicating %s, %v", footerPaktTmpl.filename, err)
	}

	return nil
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
