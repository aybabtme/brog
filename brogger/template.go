package brogger

import (
	"html/template"
)

// Base templates
const (
	applicationTmpl = "application.gohtml"
	indexTmpl       = "index.gohtml"
	styleTmpl       = "style.gohtml"
	javascriptTmpl  = "javascript.gohtml"
	headerTmpl      = "header.gohtml"
	footerTmpl      = "footer.gohtml"
)

type TemplateManager struct {
	brog      *Brog
	path      string
	templates map[string]*template.Template
}

func StartTemplateManager(brog *Brog, templPath string) (*TemplateManager, error) {

	tmpMngr := &TemplateManager{
		brog:      brog,
		path:      templPath,
		templates: make(map[string]*template.Template),
	}

	// By default, load the bin2go strings since they're compiled with Brog.
	// When a file watch changes, replace the bin2go string with the file

	brog.Warn("Not implemented yet! Watch path at %s", templPath)

	return tmpMngr, nil
}

func (t *TemplateManager) loadFromFile(filename string) error {
	t.brog.Warn("Not implemented yet! Load template from file %s", filename)
	return nil
}
