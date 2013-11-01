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
	applicationTmpl = packed{"application.gohtml", base_templates_application_gohtml}
	indexTmpl       = packed{"index.gohtml", base_templates_index_gohtml}
	styleTmpl       = packed{"style.gohtml", base_templates_style_gohtml}
	javascriptTmpl  = packed{"javascript.gohtml", base_templates_javascript_gohtml}
	headerTmpl      = packed{"header.gohtml", base_templates_header_gohtml}
	footerTmpl      = packed{"footer.gohtml", base_templates_footer_gohtml}
)

// Base CSS
var brogCss = packed{cssPath + "brog.css", base_assets_css_brog_css}

// Base JS
var brogJs = packed{jsPath + "brog.js", base_assets_js_brog_js}

// Base posts
var samplePost = packed{"sample.md", base_posts_sample_md}

func CopyBrogBinaries(conf *Config) error {
	os.MkdirAll(conf.PostPath, 0740)
	os.MkdirAll(conf.AssetPath, 0740)
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

	if err := brogCss.ReplicateInDir(conf.AssetPath); err != nil {
		return fmt.Errorf("replicating %s, %v", brogCss.filename, err)
	}
	if err := brogJs.ReplicateInDir(conf.AssetPath); err != nil {
		return fmt.Errorf("replicating %s, %v", brogJs.filename, err)
	}
	if err := samplePost.ReplicateInDir(conf.AssetPath); err != nil {
		return fmt.Errorf("replicating %s, %v", samplePost.filename, err)
	}

	return nil
}
