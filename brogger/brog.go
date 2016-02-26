package brogger

import (
	"compress/gzip"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path"
	"runtime"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/aybabtme/log"
	"github.com/prometheus/client_golang/prometheus"
)

// Brog loads its configuration file, provide logging facility, serves
// brog posts and watches for changes in posts and templates.
type Brog struct {
	isProd      bool
	Config      *Config
	netList     net.Listener
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

	brog := &Brog{
		Config: config,
		isProd: isProd,
	}

	brog.sigCatch()

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

	if len(errs) != 0 {
		return fmt.Errorf("caught errors while closing, %v", errs)
	}

	if b.netList != nil {
		errHandler(b.netList.Close())
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
		log.KV("addr", addr).Info("listening on TCP")
	} else {
		addr = ""
		log.KV("sock", "unix://"+sock).Info("listening on UNIX socket")
	}

	log.Info("brog is starting")

	if err := b.startWatchers(); err != nil {
		return fmt.Errorf("starting watchers, %v", err)
	}

	// don't add middleware to prometheus/heartbeat
	b.HandleFunc("/debug/metrics", b.prometheusHandler(prometheus.Handler().ServeHTTP, "srv", "metrics"))
	b.HandleFunc("/heartbeat", b.prometheusHandler(b.heartBeat, "srv", "heartbeat"))
	b.middlewares = append(b.middlewares, b.logHandlerFunc)

	// langSelect shouldn't have language middleware on it
	b.HandleFunc("/changelang", b.prometheusHandler(b.langSelectFunc, "srv", "changelang"))
	b.middlewares = append(b.middlewares, b.langHandlerFunc)

	b.HandleFunc("/posts/", b.prometheusHandler(b.postFunc, "srv", "posts"))
	b.HandleFunc("/pages/", b.prometheusHandler(b.pageFunc, "srv", "pages"))
	b.HandleFunc("/", b.prometheusHandler(b.indexFunc, "srv", "all"))

	fileServer := http.FileServer(http.Dir(b.Config.AssetPath))
	http.Handle("/assets/", http.StripPrefix("/assets/",
		b.prometheusHandler(
			b.logHandlerFunc(b.gzipHandler(fileServer)),
			"srv", "assets",
		),
	))

	if addr != "" {
		b.netList, err = net.Listen("tcp", addr)
	} else {
		b.netList, err = net.Listen("unix", sock)
	}
	if err != nil {
		return err
	}

	log.Info("brog is ready")

	return http.Serve(b.netList, nil)
}

////////////////////////////////////////////////////////////////////////////////
// Helpers
////////////////////////////////////////////////////////////////////////////////

func (b *Brog) startWatchers() error {
	log.Info("starting file watches")

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

// Make sure we are going to catch interupts
func (b *Brog) sigCatch() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		<-c
		log.Info("caught signal, stopping")
		if err := b.Close(); err != nil {
			log.Err(err).Fatal("failed to close cleanly")
		}
		os.Exit(0)
	}()
}

// Logging helpers

func (b *Brog) logHandlerFunc(h http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {

		ll := log.
			KV("req.raddr", req.RemoteAddr).
			KV("req.uri", req.RequestURI).
			KV("req.ua", req.UserAgent())
		ll.Info("begin request")

		now := time.Now()
		h.ServeHTTP(w, req)
		ll.KV("req.dur_ms", time.Since(now).Nanoseconds()/1e6).Info("done request")
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

//gzip handler for the assets files
func (b *Brog) gzipHandler(h http.Handler) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if !strings.Contains(req.Header.Get("Accept-Encoding"), "gzip") {
			h.ServeHTTP(w, req)
			return
		}
		w.Header().Set("Content-Encoding", "gzip")
		gz := gzip.NewWriter(w)
		defer func() {
			_ = gz.Close()
		}()
		gzrw := gzipResponseWriter{Writer: gz, ResponseWriter: w}
		h.ServeHTTP(gzrw, req)
	})
}

// prometheus handler for matrics
func (b *Brog) prometheusHandler(h http.HandlerFunc, k, v string) http.HandlerFunc {
	return http.HandlerFunc(
		prometheus.InstrumentHandlerFuncWithOpts(prometheus.SummaryOpts{
			Namespace:   "brog",
			Subsystem:   "http",
			Objectives:  map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001, 0.999: 0.0001},
			ConstLabels: map[string]string{k: v},
		}, h),
	)
}

////////////////////////////////////////////////////////////////////////////////
// HandlerFuncs
////////////////////////////////////////////////////////////////////////////////

// heartBeat answers 200 to any request.
func (b *Brog) heartBeat(rw http.ResponseWriter, req *http.Request) {
	log.Info("hearbeat!")
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
			log.Err(err).Error("couldn't render index template")
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
			log.Err(err).KV("post.id", postID).Error("couldn't render post template")
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
			log.Err(err).KV("page.id", pageID).Error("couldn't render page template")
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
		if err := t.Execute(rw, data); err != nil {
			log.Err(err).Error("couldn't render language selection template")
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
