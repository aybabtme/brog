package brogger

import (
	"fmt"
	"github.com/howeyc/fsnotify"
	"path"
	"path/filepath"
	"strings"
	"sync"
	// Using `text` instead of `html` to keep the
	// JS/CSS/HTML from inside Markdown posts
	"text/template"
)

const (
	appTmplName    = "application.gohtml"
	indexTmplName  = "index.gohtml"
	postTmplName   = "post.gohtml"
	styleTmplName  = "style.gohtml"
	jsTmplName     = "javascript.gohtml"
	headerTmplName = "header.gohtml"
	footerTmplName = "footer.gohtml"
)

type TemplateManager struct {
	brog *Brog
	path string

	watcher *fsnotify.Watcher // Listens on `path`
	die     chan struct{}     // To kill the watcher goroutine

	mu    sync.RWMutex // Locks the templates
	index *template.Template
	post  *template.Template
}

func StartTemplateManager(brog *Brog, templPath string) (*TemplateManager, error) {

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("getting template watcher, %v", err)
	}

	tmpMngr := &TemplateManager{
		brog:    brog,
		path:    templPath,
		watcher: watcher,
		die:     make(chan struct{}),
		mu:      sync.RWMutex{},
	}

	if err := tmpMngr.initializeAppTmpl(); err != nil {
		return nil, fmt.Errorf("initializing templates, %v", err)
	}

	tmpMngr.watchForChanges(templPath)

	return tmpMngr, nil
}

func (t *TemplateManager) PostTmpl() template.Template {
	t.mu.RLock()
	defer t.mu.RUnlock()
	// Give a copy
	return *t.post
}

func (t *TemplateManager) IndexTmpl() template.Template {
	t.mu.RLock()
	defer t.mu.RUnlock()
	// Give a copy
	return *t.index
}

func (t *TemplateManager) Close() error {
	t.die <- struct{}{}
	return t.watcher.Close()
}

func (t *TemplateManager) initializeAppTmpl() error {

	prefix := func(filename string) string {
		return path.Join(t.brog.Config.TemplatePath, filename)
	}

	index, err := template.ParseFiles(
		prefix(appTmplName),
		prefix(styleTmplName),
		prefix(jsTmplName),
		prefix(headerTmplName),
		prefix(footerTmplName),
		prefix(indexTmplName),
	)
	if err != nil {
		return fmt.Errorf("parsing index template at '%s', %v", prefix(indexTmplName), err)
	}
	post, err := template.ParseFiles(
		prefix(appTmplName),
		prefix(styleTmplName),
		prefix(jsTmplName),
		prefix(headerTmplName),
		prefix(footerTmplName),
		prefix(postTmplName),
	)
	if err != nil {
		return fmt.Errorf("parsing post template at '%s', %v", prefix(postTmplName), err)
	}

	t.mu.Lock()
	t.index = index
	t.post = post
	t.mu.Unlock()

	return nil
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
		t.brog.Debug("Templates ignore files in '%s': %s", ext, ev.Name)
	}

	if ev.IsModify() {
		t.brog.Debug("Template '%s' changed, parsing templates again", ev.Name)
		if err := t.initializeAppTmpl(); err != nil {
			t.brog.Err("Failed reinitialization of templates, %v", err)
		}
		return
	}

	if ev.IsCreate() || ev.IsRename() || ev.IsDelete() {
		t.brog.Err("Not yet implemented, %s", ev.String())
		return
	}

	t.brog.Err("FileEvent '%s' is not recognized", ev.String())
}

func stripExtension(fullpath string) string {
	filename := filepath.Base(fullpath)
	extLen := len(path.Ext(filename))
	// Strip the file extension
	return filename[:len(filename)-extLen]
}
