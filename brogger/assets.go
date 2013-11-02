package brogger

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
)

const (
	jsPath  = "js/"
	cssPath = "css/"
)

type packed struct {
	filename string
	data     []byte
}

func (p *packed) ReplicateInDir(dirpath string) error {
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

func CopyBrogBinaries(conf *Config) error {

	// Posts
	os.MkdirAll(conf.PostPath, 0740)
	if err := samplePost.ReplicateInDir(conf.PostPath); err != nil {
		return fmt.Errorf("replicating %s, %v", samplePost.filename, err)
	}

	// Assets
	os.MkdirAll(path.Join(conf.AssetPath, cssPath), 0740)
	os.MkdirAll(path.Join(conf.AssetPath, jsPath), 0740)
	if err := brogCss.ReplicateInDir(conf.AssetPath); err != nil {
		return fmt.Errorf("replicating %s, %v", brogCss.filename, err)
	}
	if err := brogJs.ReplicateInDir(conf.AssetPath); err != nil {
		return fmt.Errorf("replicating %s, %v", brogJs.filename, err)
	}

	// Templates
	os.MkdirAll(conf.TemplatePath, 0740)
	if err := appPaktTmpl.ReplicateInDir(conf.TemplatePath); err != nil {
		return fmt.Errorf("replicating %s, %v", appPaktTmpl.filename, err)
	}
	if err := indexPaktTmpl.ReplicateInDir(conf.TemplatePath); err != nil {
		return fmt.Errorf("replicating %s, %v", indexPaktTmpl.filename, err)
	}
	if err := postPaktTmpl.ReplicateInDir(conf.TemplatePath); err != nil {
		return fmt.Errorf("replicating %s, %v", postPaktTmpl.filename, err)
	}
	if err := stylePaktTmpl.ReplicateInDir(conf.TemplatePath); err != nil {
		return fmt.Errorf("replicating %s, %v", stylePaktTmpl.filename, err)
	}
	if err := javascriptPaktTmpl.ReplicateInDir(conf.TemplatePath); err != nil {
		return fmt.Errorf("replicating %s, %v", javascriptPaktTmpl.filename, err)
	}
	if err := headerPaktTmpl.ReplicateInDir(conf.TemplatePath); err != nil {
		return fmt.Errorf("replicating %s, %v", headerPaktTmpl.filename, err)
	}
	if err := footerPaktTmpl.ReplicateInDir(conf.TemplatePath); err != nil {
		return fmt.Errorf("replicating %s, %v", footerPaktTmpl.filename, err)
	}

	return nil
}

func CopyBlankToFilename(conf *Config, filename string) error {
	fullpath := path.Clean(conf.PostPath) + string(os.PathSeparator) + filename
	return ioutil.WriteFile(fullpath, basePostsBlankMd, 0640)
}
