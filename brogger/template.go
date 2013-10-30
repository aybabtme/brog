package brogger

import (
	"fmt"
	"html/template"
	"io/ioutil"
	"os"
)

type TemplateManager struct {
	brog      *Brog
	path      string
	templates map[string]*template.Template
}

func StartTemplateManager(brog *Brog, templPath string) (*TemplateManager, error) {

	os.MkdirAll(templPath, 0740)
	fileInfos, err := ioutil.ReadDir(templPath)
	if err != nil {
		return nil, fmt.Errorf("listing directory '%s', %v", templPath, err)
	}

	tmpMngr := &TemplateManager{
		brog:      brog,
		path:      templPath,
		templates: make(map[string]*template.Template),
	}

	for _, fileInfo := range fileInfos {
		if !fileInfo.IsDir() {
			err := tmpMngr.loadFromFile(fileInfo.Name())
			if err != nil {
				brog.Warn("Loading template from file '%s' failed, %v", fileInfo.Name(), err)
			}
		}
	}

	brog.Warn("Not implemented yet! Watch path at %s", templPath)

	return tmpMngr, nil
}

func (t *TemplateManager) loadFromFile(filename string) error {
	t.brog.Warn("Not implemented yet! Load template from file %s", filename)
	return nil
}
