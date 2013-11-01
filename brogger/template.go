package brogger

import (
	"path"
	// Using `text` instead of `html` to keep the
	// JS/CSS/HTML from inside Markdown posts
	"text/template"
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
	tmpMngr.registerPackedTemplate(applicationTmpl)
	tmpMngr.registerPackedTemplate(indexTmpl)
	tmpMngr.registerPackedTemplate(styleTmpl)
	tmpMngr.registerPackedTemplate(javascriptTmpl)
	tmpMngr.registerPackedTemplate(headerTmpl)
	tmpMngr.registerPackedTemplate(footerTmpl)

	brog.Warn("Not implemented yet! Watch path at %s", templPath)

	return tmpMngr, nil
}

func (t *TemplateManager) registerPackedTemplate(pak packed) {
	name := stripExtension(pak.filename)
	tmpl := template.Must(template.New(name).Parse(string(pak.data)))
	t.templates[name] = tmpl
}

func (t *TemplateManager) loadFromFile(filename string) error {
	t.brog.Warn("Not implemented yet! Load template from file %s", filename)
	return nil
}

func stripExtension(filename string) string {
	extLen := len(path.Ext(filename))
	// Strip the file extension
	return filename[:extLen]
}
