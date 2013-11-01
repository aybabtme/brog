package brogger

import (
	"fmt"
	"github.com/howeyc/fsnotify"
	"io/ioutil"
	"path"
	"path/filepath"
	"strings"
	"sync"
	// Using `text` instead of `html` to keep the
	// JS/CSS/HTML from inside Markdown posts
	"text/template"
)

type TemplateManager struct {
	brog *Brog
	path string

	watcher *fsnotify.Watcher // Listens on `path`
	die     chan struct{}     // To kill the watcher goroutine

	mu        sync.RWMutex // Locks the `templates`
	templates map[string]*template.Template
}

func StartTemplateManager(brog *Brog, templPath string) (*TemplateManager, error) {

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("getting template watcher, %v", err)
	}

	tmpMngr := &TemplateManager{
		brog:      brog,
		path:      templPath,
		watcher:   watcher,
		die:       make(chan struct{}),
		mu:        sync.RWMutex{},
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

	tmpMngr.watchForChanges(templPath)

	return tmpMngr, nil
}

func (t *TemplateManager) GetTmpl(name string) (*template.Template, bool) {
	t.mu.RLock()
	defer t.mu.RUnlock()
	tmpl, ok := t.templates[name]
	return tmpl, ok
}

func (t *TemplateManager) SetTmpl(tmpl *template.Template) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.templates[tmpl.Name()] = tmpl
}

func (t *TemplateManager) DeleteTmpl(name string) (*template.Template, bool) {
	t.mu.Lock()
	tmpl, ok := t.templates[name]
	if !ok {
		t.mu.Unlock()
		return nil, ok
	}
	delete(t.templates, name)
	t.mu.Unlock()
	return tmpl, ok
}

func (t *TemplateManager) Close() error {
	t.die <- struct{}{}
	return t.watcher.Close()
}

func (t *TemplateManager) registerPackedTemplate(pak packed) {
	name := stripExtension(pak.filename)
	tmpl := template.Must(template.New(name).Parse(string(pak.data)))
	t.SetTmpl(tmpl)
	t.brog.Ok("Packaged template '%s' is now registered", tmpl.Name())
}

func (t *TemplateManager) watchForChanges(dirname string) error {
	go func() {
		for {
			select {
			case ev := <-t.watcher.Event:
				t.processTemplateEvent(ev)
			case err := <-t.watcher.Error:
				t.brog.Err("watching templates in '%s', %v", dirname, err)
			case <-t.die:
				return
			}
		}
	}()

	return t.watcher.Watch(dirname)
}

func (t *TemplateManager) processTemplateEvent(ev *fsnotify.FileEvent) {
	ext := strings.ToLower(filepath.Ext(ev.Name))
	switch ext {
	case ".gohtml":
	case ".tmpl":
	default:
		t.brog.Ok("Templates ignore files in '%s': %s", ext, ev.Name)
	}

	if ev.IsCreate() {
		t.processTemplateCreate(ev)
		return
	}

	if ev.IsModify() {
		t.processTemplateModify(ev)
		return
	}

	if ev.IsRename() {
		t.processTemplateRename(ev)
		return
	}

	if ev.IsDelete() {
		t.processTemplateDelete(ev)
		return
	}

	t.brog.Err("FileEvent '%s' is not recognized", ev.String())
}

func (t *TemplateManager) processTemplateRename(ev *fsnotify.FileEvent) {

	tmpl, ok := t.DeleteTmpl(stripExtension(ev.Name))

	if !ok {
		t.brog.Warn("Renamed unknown file '%s', ignoring", ev.Name)
		return
	}

	t.brog.Ok("Template '%s': old filename '%s', deleting, %d templates total",
		tmpl.Name(), ev.Name, len(t.templates))

	return
}

func (t *TemplateManager) processTemplateDelete(ev *fsnotify.FileEvent) {

	tmpl, ok := t.DeleteTmpl(stripExtension(ev.Name))

	if !ok {
		t.brog.Warn("Deleting unknown file '%s', ignoring", ev.Name)
		return
	}

	t.brog.Ok("Removing template '%s', %d templates left", tmpl.Name, len(t.templates))
	return
}

func (t *TemplateManager) processTemplateCreate(ev *fsnotify.FileEvent) {
	t.brog.Ok("New file '%s'", ev.Name)
	err := t.loadFromFile(ev.Name)
	if err != nil {
		t.brog.Err("Error loading new template at '%s', %v", ev.Name, err)
	}
}

func (t *TemplateManager) processTemplateModify(ev *fsnotify.FileEvent) {
	t.brog.Ok("Modified file '%s'", ev.Name)

	tmplName := stripExtension(ev.Name)

	tmpl, ok := t.GetTmpl(tmplName)

	if !ok {
		t.brog.Warn("File '%s' was unknown", ev.Name)
	}

	err := t.loadFromFile(ev.Name)
	if err != nil {
		t.brog.Err("Error loading new template at '%s', %v", ev.Name, err)
		if ok {

			t.DeleteTmpl(tmplName)

			t.brog.Warn("Removing related template '%s', %d templates left",
				tmpl.Name, len(t.templates))

		}
	}
}

func (t *TemplateManager) loadFromFile(filename string) error {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("loading template from file '%s', %v", filename, err)
	}

	tmplName := stripExtension(filename)
	tmpl, err := template.New(tmplName).Parse(string(data))
	if err != nil {
		return fmt.Errorf("parsing template from file '%s', %v", filename, err)
	}

	t.SetTmpl(tmpl)

	t.brog.Ok("Loaded template '%s' from file '%s', %d templates total", tmplName, filename, len(t.templates))

	return nil
}

func stripExtension(fullpath string) string {
	filename := filepath.Base(fullpath)
	extLen := len(path.Ext(filename))
	// Strip the file extension
	return filename[:len(filename)-extLen]
}
