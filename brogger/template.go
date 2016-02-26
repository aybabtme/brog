package brogger

import (
	"fmt"
	"path/filepath"
	"strings"
	"sync"

	"github.com/aybabtme/log"
	"github.com/howeyc/fsnotify"

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
		return filepath.Join(t.brog.Config.TemplatePath, filename)
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
				log.Err(err).KV("dir.name", dirname).Error("error watching templates")
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
		log.KV("ext", ext).KV("file.name", ev.Name).Info("ignoring file")
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

	log.KV("file.event", ev.String()).Error("unknown file event")
}

func (t *templateManager) processTemplateModify(ev *fsnotify.FileEvent) {
	filename := filepath.Base(ev.Name)
	tmpl, ok := allTemplates[filename]
	if !ok {
		log.KV("file.name", ev.Name).Error("ignoring, brog can only use its default templates")
		return
	}

	ll := log.KV("file.name", ev.Name)
	ll.Info("template name changed, parsing templates again")
	err := t.initializeAppTmpl()
	if err == nil {
		ll.Info("new templates have been assimilated")
		return
	}
	ll.Err(err).Error("failed to reinitialize templates, reconstructing")

	if !t.brog.Config.RewriteInvalid {
		// Nothing to do, just fail
		return
	}

	if err := tmpl.rewriteInDir(t.brog.Config.TemplatePath); err != nil {
		ll.Error("reconstruction failed")
		return
	}

	if err := t.initializeAppTmpl(); err != nil {
		ll.Err(err).Error("failed to reload templates after reconstruction")
		return
	}

	ll.Info("threat to template has been eradicated. Resistance is futile.  You will be assimilated")
}

func (t *templateManager) processTemplateDelete(ev *fsnotify.FileEvent) {

	filename := filepath.Base(ev.Name)
	tmpl, ok := allTemplates[filename]
	if !ok {
		// Don't care
		return
	}
	ll := log.KV("file.name", ev.Name)
	ll.Error("detected the destruction of a vital part")

	if !t.brog.Config.RewriteMissing {
		// Nothing to do, just fail
		return
	}

	ll.Error("reconstructing missing part")

	if err := tmpl.replicateInDir(t.brog.Config.TemplatePath); err != nil {
		ll.Err(err).Error("reconstruction failed")
		return
	}

	if err := t.initializeAppTmpl(); err != nil {
		ll.Err(err).Error("failed to reload templates after reconstruction")
		return
	}

	ll.Info("threat to template has been eradicated. Resistance is futile.  You will be assimilated")
}

func stripExtension(fullpath string) string {
	filename := filepath.Base(fullpath)
	extLen := len(filepath.Ext(filename))
	// Strip the file extension
	return filename[:len(filename)-extLen]
}
