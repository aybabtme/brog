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
	applicationTmpl = packed{"application.gohtml", baseTemplatesApplicationGohtml}
	indexTmpl       = packed{"index.gohtml", baseTemplatesIndexGohtml}
	styleTmpl       = packed{"style.gohtml", baseTemplatesStyleGohtml}
	javascriptTmpl  = packed{"javascript.gohtml", baseTemplatesJavascriptGohtml}
	headerTmpl      = packed{"header.gohtml", baseTemplatesHeaderGohtml}
	footerTmpl      = packed{"footer.gohtml", baseTemplatesFooterGohtml}
)

// Base CSS
var brogCss = packed{cssPath + "brog.css", baseAssetsCssBrogCss}

// Base JS
var brogJs = packed{jsPath + "brog.js", baseAssetsJsBrogJs}

// Base posts
var samplePost = packed{"sample.md", basePostsSampleMd}

func CopyBrogBinaries(conf *Config) error {

	// Posts
	os.MkdirAll(conf.PostPath, 0740)
	if err := samplePost.ReplicateInDir(conf.PostPath); err != nil {
		return fmt.Errorf("replicating %s, %v", samplePost.filename, err)
	}

	// Assets
	os.MkdirAll(conf.AssetPath, 0740)
	if err := brogCss.ReplicateInDir(conf.AssetPath); err != nil {
		return fmt.Errorf("replicating %s, %v", brogCss.filename, err)
	}
	if err := brogJs.ReplicateInDir(conf.AssetPath); err != nil {
		return fmt.Errorf("replicating %s, %v", brogJs.filename, err)
	}

	// Templates
	os.MkdirAll(conf.TemplatePath, 0740)
	if err := applicationTmpl.ReplicateInDir(conf.TemplatePath); err != nil {
		return fmt.Errorf("replicating %s, %v", applicationTmpl.filename, err)
	}
	if err := indexTmpl.ReplicateInDir(conf.TemplatePath); err != nil {
		return fmt.Errorf("replicating %s, %v", indexTmpl.filename, err)
	}
	if err := styleTmpl.ReplicateInDir(conf.TemplatePath); err != nil {
		return fmt.Errorf("replicating %s, %v", styleTmpl.filename, err)
	}
	if err := javascriptTmpl.ReplicateInDir(conf.TemplatePath); err != nil {
		return fmt.Errorf("replicating %s, %v", javascriptTmpl.filename, err)
	}
	if err := headerTmpl.ReplicateInDir(conf.TemplatePath); err != nil {
		return fmt.Errorf("replicating %s, %v", headerTmpl.filename, err)
	}
	if err := footerTmpl.ReplicateInDir(conf.TemplatePath); err != nil {
		return fmt.Errorf("replicating %s, %v", footerTmpl.filename, err)
	}

	return nil
}
