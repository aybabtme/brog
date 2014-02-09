package brogger

import (
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"path"
	"runtime"
	"strconv"
	"strings"
	"text/template"
	"time"
)

// Brog loads its configuration file, provide logging facility, serves
// brog posts and watches for changes in posts and templates.
type Brog struct {
	*logMux
	isProd      bool
	Config      *Config
	Pid         int
	tmplMngr    *templateManager
	postMngr    *postManager
	pageMngr    *postManager
	middlewares [](func(http.HandlerFunc) http.HandlerFunc)
}

type appContent struct {
	Posts     []*post
	Pages     []*post
	Languages []string
	CurPost   *post
	Redir     string
}

////////////////////////////////////////////////////////////////////////////////
// Brog exported funcs
////////////////////////////////////////////////////////////////////////////////

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
		Pid:    os.Getpid(),
	}

	if isProd {
		if err := brog.writePID(); err != nil {
			panic(err)
		}
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

	if b.pageMngr != nil {
		errHandler(b.pageMngr.Close())
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

	var sock string
	if b.isProd {
		sock = b.Config.ProdPort
	} else {
		sock = b.Config.DevelPort
	}
	port, err := strconv.ParseInt(sock, 10, 0)

	var addr string
	if err == nil {
		addr = fmt.Sprintf("%s:%d", b.Config.Hostname, port)
		b.Ok("CAPTAIN: Open channel, %s", addr)
	} else {
		addr = ""
		b.Ok("CAPTAIN: Open channel, unix://%s", sock)
	}

	b.Warn("ON SCREEN: We are the Brog. Resistance is futile.")

	if err := b.startWatchers(); err != nil {
		return fmt.Errorf("starting watchers, %v", err)
	}

	b.HandleFunc("/heartbeat", b.heartBeat) // don't add middleware to heartbeat
	b.middlewares = append(b.middlewares, b.logHandlerFunc)

	// langSelect shouldn't have language middleware on it
	b.HandleFunc("/changelang", b.langSelectFunc)
	b.middlewares = append(b.middlewares, b.langHandlerFunc)

	b.HandleFunc("/posts/", b.postFunc)
	b.HandleFunc("/pages/", b.pageFunc)
	b.HandleFunc("/", b.indexFunc)

	fileServer := http.FileServer(http.Dir(b.Config.AssetPath))
	http.Handle("/assets/", http.StripPrefix("/assets/", b.logHandler(fileServer)))

	b.Ok("Assimilation completed.")
	if b.isProd {
		b.Warn("Going live in production.")
	}
	var l net.Listener
	if addr != "" {
		l, err = net.Listen("tcp", addr)
	} else {
		l, err = net.Listen("unix", sock)
	}
	if err != nil {
		return err
	}
	return http.Serve(l, nil)
}

////////////////////////////////////////////////////////////////////////////////
// Helpers
////////////////////////////////////////////////////////////////////////////////

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

	pageMngr, err := startPostManager(b, b.Config.PagePath)
	if err != nil {
		return fmt.Errorf("starting page manager, %v", err)
	}

	b.postMngr = postMngr
	b.pageMngr = pageMngr

	return nil
}

// write PID because sysadmin
func (b *Brog) writePID() error {
	b.Ok("Galactic coordinates: %d,%02d", b.Pid/100, b.Pid%100)

	pidBytes := []byte(strconv.Itoa(b.Pid))
	if err := ioutil.WriteFile("brog.pid", pidBytes, 0755); err != nil {
		return fmt.Errorf("error writing to PID file: %v", err)
	}

	return nil
}

// Logging helpers

func (b *Brog) logHandlerFunc(h http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {

		b.Ok("request by %s for '%s' as '%s'",
			req.RemoteAddr, req.RequestURI, req.UserAgent())

		now := time.Now()
		h.ServeHTTP(w, req)
		b.Ok("Done in %s", time.Since(now))
	})
}

func (b *Brog) logHandler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {

		b.Ok("request by %s for '%s' as '%s'",
			req.RemoteAddr, req.RequestURI, req.UserAgent())

		now := time.Now()
		h.ServeHTTP(w, req)
		b.Ok("Done in %s", time.Since(now))
	})
}

////////////////////////////////////////////////////////////////////////////////
// Middleware
////////////////////////////////////////////////////////////////////////////////

func (b *Brog) HandleFunc(path string, h http.HandlerFunc) {
	for _, middleware := range b.middlewares {
		h = middleware(h)
	}
	http.HandleFunc(path, h)
}

////////////////////////////////////////////////////////////////////////////////
// HandlerFuncs
////////////////////////////////////////////////////////////////////////////////

// heartBeat answers 200 to any request.
func (b *Brog) heartBeat(rw http.ResponseWriter, req *http.Request) {
	b.Debug("Hearbeat!")
	rw.WriteHeader(http.StatusOK)
}

func (b *Brog) indexFunc(rw http.ResponseWriter, req *http.Request) {

	lang, _ := b.extractLanguage(req)
	pages := b.pageMngr.GetAllPostsWithLanguage(lang)
	posts := b.postMngr.GetAllPostsWithLanguage(lang)

	data := appContent{
		Posts:     posts,
		Pages:     pages,
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

func (b *Brog) postFunc(rw http.ResponseWriter, req *http.Request) {

	lang, _ := b.extractLanguage(req)
	pages := b.pageMngr.GetAllPostsWithLanguage(lang)

	postID := path.Base(strings.SplitN(req.RequestURI, "?", 2)[0])
	post, ok := b.postMngr.GetPost(postID)
	if !ok {
		b.Warn("not found, %v", req)
		http.NotFound(rw, req)
		return
	}

	data := appContent{
		Posts:     nil,
		Pages:     pages,
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

func (b *Brog) pageFunc(rw http.ResponseWriter, req *http.Request) {

	lang, _ := b.extractLanguage(req)
	pages := b.pageMngr.GetAllPostsWithLanguage(lang)

	pageID := path.Base(strings.SplitN(req.RequestURI, "?", 2)[0])
	page, ok := b.pageMngr.GetPost(pageID)

	if !ok {
		b.Warn("not found, %v", req)
		http.NotFound(rw, req)
		return
	}

	data := appContent{
		Posts:     nil,
		Pages:     pages,
		Languages: b.Config.Languages,
		CurPost:   page,
	}

	b.tmplMngr.DoWithPost(func(t *template.Template) {
		if err := t.Execute(rw, data); err != nil {
			b.Err("serving post request for ID=%s, %v", pageID, err)
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}
	})
}

////////////////////////////////////////////////////////////////////////////////
// Multilingual support
////////////////////////////////////////////////////////////////////////////////

func (b *Brog) validLangInQuery(lang string) bool {
	for _, val := range b.Config.Languages {
		if strings.Contains(lang, val) {
			return true
		}
	}
	return false
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
		rw.Header().Add("Set-Cookie", "lang="+lang+";Path=/")
	}
}

func (b *Brog) langSelectFunc(rw http.ResponseWriter, req *http.Request) {
	b.Debug("Language not set for multilingual blog")

	redirpath := req.URL.Path
	if redirpath == "/changelang" {
		redirpath = "/"
	}

	data := appContent{
		Posts:     nil,
		Pages:     nil,
		Languages: b.Config.Languages,
		CurPost:   nil,
		Redir:     redirpath,
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

func (b *Brog) langHandlerFunc(h http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		if b.Config.Multilingual {
			_, validLang := b.extractLanguage(req)
			b.setLangCookie(req, rw)
			if !validLang {
				b.langSelectFunc(rw, req)
				return
			}
		}
		h.ServeHTTP(rw, req)
	})
}
