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
	appTmplName        = "application.gohtml"
	indexTmplName      = "index.gohtml"
	postTmplName       = "post.gohtml"
	langSelectTmplName = "langselect.gohtml"
	styleTmplName      = "style.gohtml"
	jsTmplName         = "javascript.gohtml"
	headerTmplName     = "header.gohtml"
	footerTmplName     = "footer.gohtml"
)

type templateManager struct {
	brog *Brog
	path string

	watcher *fsnotify.Watcher // Listens on `path`
	die     chan struct{}     // To kill the watcher goroutine

	mu         sync.RWMutex // Locks the templates
	index      *template.Template
	post       *template.Template
	langselect *template.Template
}

func startTemplateManager(brog *Brog, templPath string) (*templateManager, error) {

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("getting template watcher, %v", err)
	}

	tmpMngr := &templateManager{
		brog:    brog,
		path:    templPath,
		watcher: watcher,
		die:     make(chan struct{}),
		mu:      sync.RWMutex{},
	}

	if err := tmpMngr.initializeAppTmpl(); err != nil {
		return nil, fmt.Errorf("initializing templates, %v", err)
	}

	if err := tmpMngr.watchForChanges(templPath); err != nil {
		return nil, fmt.Errorf("starting watch for changes on '%s', %v", templPath, err)
	}

	return tmpMngr, nil
}

func (t *templateManager) DoWithPost(do func(*template.Template)) {
	t.mu.RLock()
	defer t.mu.RUnlock()
	do(t.post)
}

func (t *templateManager) DoWithIndex(do func(*template.Template)) {
	t.mu.RLock()
	defer t.mu.RUnlock()
	do(t.index)
}

func (t *templateManager) DoWithLangSelect(do func(*template.Template)) {
	t.mu.RLock()
	defer t.mu.RUnlock()
	do(t.langselect)
}

func (t *templateManager) Close() error {
	t.die <- struct{}{}
	return t.watcher.Close()
}

func (t *templateManager) initializeAppTmpl() error {

	prefix := func(filename string) string {
		return path.Join(t.brog.Config.TemplatePath, filename)
	}

	indexApp, err := template.ParseFiles(
		prefix(appTmplName),
		prefix(styleTmplName),
		prefix(jsTmplName),
		prefix(headerTmplName),
		prefix(footerTmplName),
	)
	if err != nil {
		return fmt.Errorf("parsing indexApp template, %v", err)
	}
	index, err := indexApp.ParseFiles(prefix(indexTmplName))
	if err != nil {
		return fmt.Errorf("parsing index template at '%s', %v", prefix(indexTmplName), err)
	}
	postApp, err := template.ParseFiles(
		prefix(appTmplName),
		prefix(styleTmplName),
		prefix(jsTmplName),
		prefix(headerTmplName),
		prefix(footerTmplName),
	)
	if err != nil {
		return fmt.Errorf("parsing postApp template, %v", err)
	}
	post, err := postApp.ParseFiles(prefix(postTmplName))
	if err != nil {
		return fmt.Errorf("parsing post template at '%s', %v", prefix(postTmplName), err)
	}

	langSelectApp, err := template.ParseFiles(
		prefix(appTmplName),
		prefix(styleTmplName),
		prefix(jsTmplName),
		prefix(headerTmplName),
		prefix(footerTmplName),
	)
	if err != nil {
		return fmt.Errorf("parsing langSelectApp template, %v", err)
	}
	langSelect, err := langSelectApp.ParseFiles(prefix(langSelectTmplName))
	if err != nil {
		return fmt.Errorf("parsing langSelect template at '%s', %v", prefix(langSelectTmplName), err)
	}

	t.mu.Lock()
	t.index = index
	t.post = post
	t.langselect = langSelect
	t.mu.Unlock()

	return nil
}

func (t *templateManager) watchForChanges(dirname string) error {
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

func (t *templateManager) processTemplateEvent(ev *fsnotify.FileEvent) {
	ext := strings.ToLower(filepath.Ext(ev.Name))
	switch ext {
	case ".gohtml":
	case ".tmpl":
	default:
		t.brog.Debug("Templates ignore files in '%s': %s", ext, ev.Name)
		return
	}

	if ev.IsCreate() {
		// Nothing to do, all usefull templates files MUST exist at this
		// time or brog would not have started.
		return
	}

	if ev.IsModify() {
		t.processTemplateModify(ev)
		return
	}

	if ev.IsRename() || ev.IsDelete() {
		t.processTemplateDelete(ev)
		return
	}

	t.brog.Err("FileEvent '%s' is not recognized", ev.String())
}

func (t *templateManager) processTemplateModify(ev *fsnotify.FileEvent) {
	filename := path.Base(ev.Name)
	tmpl, ok := allTemplates[filename]
	if !ok {
		t.brog.Watch("'%s' ignored. Brog can only use its default templates.", ev.Name)
		return
	}

	t.brog.Watch("Template '%s' changed, parsing templates again.", ev.Name)
	err := t.initializeAppTmpl()
	if err == nil {
		t.brog.Watch("Assimilation completed. '%s' has become one with the brog.", ev.Name)
		return
	}
	t.brog.Err("Failed reinitialization of templates, %v", err)

	t.brog.Err("Brog detected the corruption of a vital part: %s", ev.Name)

	if !t.brog.Config.RewriteInvalid {
		// Nothing to do, just fail
		return
	}

	t.brog.Warn("Brog to eradicate threat. Overwriting '%s'.", ev.Name)

	if err := tmpl.rewriteInDir(t.brog.Config.TemplatePath); err != nil {
		t.brog.Err("Eradication of '%s' failed, %v", filename, err)
		return
	}

	if err := t.initializeAppTmpl(); err != nil {
		t.brog.Err("Templates fail despite eradication of '%s', %v", filename, err)
		return
	}

	t.brog.Warn("Brog has successfully eradicated threat to template '%s'", ev.Name)
	t.brog.Warn("Resistance is futile.  You will be assimilated")
}

func (t *templateManager) processTemplateDelete(ev *fsnotify.FileEvent) {

	filename := path.Base(ev.Name)
	tmpl, ok := allTemplates[filename]
	if !ok {
		// Don't care
		return
	}

	t.brog.Err("Brog detected the destruction of a vital part: %s", ev.Name)

	if !t.brog.Config.RewriteMissing {
		// Nothing to do, just fail
		return
	}

	t.brog.Warn("Brog to regenerate '%s'.", ev.Name)

	if err := tmpl.replicateInDir(t.brog.Config.TemplatePath); err != nil {
		t.brog.Err("Regeneration of '%s' failed, %v", filename, err)
		return
	}

	if err := t.initializeAppTmpl(); err != nil {
		t.brog.Err("Templates fail to load despite regeneration of '%s', %v", filename, err)
		return
	}

	t.brog.Warn("Brog has successfully eradicated threat to template '%s'", ev.Name)
	t.brog.Warn("Resistance is futile.  You will be assimilated")
}

func stripExtension(fullpath string) string {
	filename := filepath.Base(fullpath)
	extLen := len(path.Ext(filename))
	// Strip the file extension
	return filename[:len(filename)-extLen]
}
