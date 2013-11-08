package brogger

import (
	"fmt"
	"net/http"
	"path/filepath"
	"runtime"
	"strings"
	"text/template"
	"time"
)

// Brog loads its configuration file, provide logging facility, serves
// brog posts and watches for changes in posts and templates.
type Brog struct {
	*logMux
	isProd   bool
	Config   *Config
	tmplMngr *templateManager
	postMngr *postManager
}

type appContent struct {
	Posts     []*post
	Languages []string
	CurPost   *post
}

// PrepareBrog creates a Brog instance, loading it's configuration from
// a file named `ConfigFilename` that must lie at the current working
// directory. It creates a logging file at the location specified in the
// config file.
// If anything goes wrong during that process, it will return an error
// explaining where it happened.
func PrepareBrog(isProd bool) (*Brog, error) {
	config, err := loadConfig()
	if err != nil {
		return nil, fmt.Errorf("preparing brog's configuration, %v", err)
	}

	logMux, err := makeLogMux(config)
	if err != nil {
		return nil, fmt.Errorf("making log multiplexer on path %s, %v", config.LogFilename, err)
	}

	brog := &Brog{
		logMux: logMux,
		Config: config,
		isProd: isProd,
	}

	return brog, nil
}

// Close ensures that all brog's resources are closed and released.
func (b *Brog) Close() error {
	errs := []error{}
	errHandler := func(err error) {
		if err != nil {
			errs = append(errs, err)
		}
	}

	if b.postMngr != nil {
		errHandler(b.postMngr.Close())
	}

	if b.tmplMngr != nil {
		errHandler(b.tmplMngr.Close())
	}

	if b.logMux != nil {
		errHandler(b.logMux.Close())
	}
	if len(errs) != 0 {
		return fmt.Errorf("caught errors while closing, %v", errs)
	}
	return nil
}

// ListenAndServe starts watching the path specified in `ConfigFilename`
// for changes and starts serving brog's content, again according to the
// settings in `ConfigFilename`.
func (b *Brog) ListenAndServe() error {

	runtime.GOMAXPROCS(b.Config.MaxCPUs)

	var port int
	if b.isProd {
		port = b.Config.ProdPortNumber
	} else {
		port = b.Config.DevelPortNumber
	}
	addr := fmt.Sprintf("%s:%d", b.Config.Hostname, port)

	b.Ok("CAPTAIN: Open channel, %s", addr)
	b.Warn("ON SCREEN: We are the Brog. Resistance is futile.")

	if err := b.startWatchers(); err != nil {
		return fmt.Errorf("starting watchers, %v", err)
	}

	http.HandleFunc("/heartbeat", b.heartBeat)
	http.HandleFunc("/changelang", b.langSelectFunc)
	http.HandleFunc("/posts/", b.postFunc)
	http.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir(b.Config.AssetPath))))
	http.HandleFunc("/", b.indexFunc)

	b.Ok("Assimilation completed.")
	if b.isProd {
		b.Warn("Going live in production.")
	}
	return http.ListenAndServe(addr, nil)
}

func (b *Brog) extractLanguage(req *http.Request) (string, bool) {
	lang := req.URL.RawQuery
	langCookie, err := req.Cookie("lang")
	if lang == "" && err == nil {
		lang = langCookie.Value
	}
	langSet := false
	for _, val := range b.Config.Languages {
		if strings.Contains(lang, val) {
			langSet = true
			break
		}
	}
	return lang, langSet
}

func (b *Brog) setLangCookie(req *http.Request, rw http.ResponseWriter) {
	lang := req.URL.RawQuery
	_, err := req.Cookie("lang")
	if lang != "" && (err != nil || strings.HasSuffix(req.Referer(), "/changelang")) {
		rw.Header().Add("Set-Cookie", "lang="+lang)
	}
}

func (b *Brog) startWatchers() error {
	b.Debug("Starting watchers")

	tmplMngr, err := startTemplateManager(b, b.Config.TemplatePath)
	if err != nil {
		return fmt.Errorf("starting template manager, %v", err)
	}
	b.tmplMngr = tmplMngr

	postMngr, err := startPostManager(b, b.Config.PostPath)
	if err != nil {
		return fmt.Errorf("starting post manager, %v", err)
	}
	b.postMngr = postMngr
	return nil
}

// heartBeat answers 200 to any request.
func (b *Brog) heartBeat(rw http.ResponseWriter, req *http.Request) {
	b.Debug("Hearbeat!")
	rw.WriteHeader(http.StatusOK)
}

func (b *Brog) indexFunc(rw http.ResponseWriter, req *http.Request) {
	defer statCount(b, req)()

	var posts []*post
	if b.Config.Multilingual {
		lang, validLang := b.extractLanguage(req)
		b.setLangCookie(req, rw)
		if !validLang {
			b.langSelectFunc(rw, req)
			return
		}
		posts = b.postMngr.GetAllPostsWithLanguage(lang)
		b.Debug("Serving %d posts with language %s to requester", len(posts), lang)
	} else {
		posts = b.postMngr.GetAllPosts()
		b.Debug("Serving %d posts to requester", len(posts))
	}

	data := appContent{
		Posts:     posts,
		Languages: b.Config.Languages,
		CurPost:   nil,
	}

	b.tmplMngr.DoWithIndex(func(t *template.Template) {
		if err := t.Execute(rw, data); err != nil {
			b.Err("serving index request, %v", err)
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}
	})
}

func (b *Brog) langSelectFunc(rw http.ResponseWriter, req *http.Request) {
	/* If the lang is already set, reset it */
	b.Debug("Language not set for multilingual blog")
	data := appContent{
		Posts:     nil,
		Languages: b.Config.Languages,
		CurPost:   nil,
	}

	b.tmplMngr.DoWithLangSelect(func(t *template.Template) {
		b.Debug("Sending language selection screen")
		if err := t.Execute(rw, data); err != nil {
			b.Err("serving index, language select, request, %v", err)
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}
	})
}

func (b *Brog) postFunc(rw http.ResponseWriter, req *http.Request) {
	defer statCount(b, req)()

	postID := filepath.Base(req.RequestURI)
	post, ok := b.postMngr.GetPost(postID)
	if !ok {
		b.Warn("not found, %v", req)
		http.NotFound(rw, req)
		return
	}

	data := appContent{
		Posts:     nil,
		Languages: b.Config.Languages,
		CurPost:   post,
	}

	b.tmplMngr.DoWithPost(func(t *template.Template) {
		if err := t.Execute(rw, data); err != nil {
			b.Err("serving post request for ID=%s, %v", postID, err)
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}
	})
}

func statCount(b *Brog, req *http.Request) func() {
	now := time.Now()
	return func() {
		b.Ok("Done in %s, req=%v", time.Since(now), req.URL)
	}
}
