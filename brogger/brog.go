package brogger

import (
	"fmt"
	"net/http"
	"runtime"
)

type Brog struct {
	*logMux
	Config   *Config
	tmplMngr *TemplateManager
	postMngr *PostManager
}

func PrepareBrog() (*Brog, error) {
	config, err := loadConfig()
	if err != nil {
		return nil, fmt.Errorf("preparing brog's configuration, %v", err)
	}

	logMux, err := makeLogMux(config.LogFilename)
	if err != nil {
		return nil, fmt.Errorf("making log multiplex on path %s, %v", config.LogFilename, err)
	}

	// Must be done before any manager is started
	CopyBrogBinaries(config)

	brog := &Brog{
		logMux: logMux,
		Config: config,
	}

	tmplMngr, err := StartTemplateManager(brog, config.TemplatePath)
	if err != nil {
		return nil, fmt.Errorf("starting template manager, %v", err)
	}
	brog.tmplMngr = tmplMngr

	postMngr, err := StartPostManager(brog, config.PostPath)
	if err != nil {
		return nil, fmt.Errorf("starting post manager, %v", err)
	}
	brog.postMngr = postMngr

	runtime.GOMAXPROCS(config.MaxCPUs)

	return brog, nil
}

// heartBeat answers 200 to any request.
func (b *Brog) heartBeat(rw http.ResponseWriter, req *http.Request) {
	rw.WriteHeader(http.StatusOK)
}

func (b *Brog) indexFunc(rw http.ResponseWriter, req *http.Request) {
	rw.WriteHeader(http.StatusOK)
	fmt.Fprint(rw, `<!doctype html>
<html>
<head><title>Hello</title></head>
<body>`)
	for _, post := range b.postMngr.GetAllPosts() {
		fmt.Fprintf(rw, "<h1>%s</h1>", post.Title)
	}
	fmt.Fprint(rw, "</body></html>")
}

func (b *Brog) postFunc(rw http.ResponseWriter, req *http.Request) {
	rw.WriteHeader(http.StatusNotImplemented)
}

func (b *Brog) ListenAndServe() error {

	http.HandleFunc("/heartbeat", b.heartBeat)

	http.HandleFunc("/", b.indexFunc)
	http.HandleFunc("/post/", b.postFunc)

	http.Handle("/assets", http.StripPrefix("/assets", http.FileServer(http.Dir("assets/"))))

	addr := fmt.Sprintf("%s:%d", b.Config.Hostname, b.Config.PortNumber)
	b.Ok("Borg open for business on %s", addr)
	return http.ListenAndServe(addr, nil)
}

func (b *Brog) Close() {
	defer b.postMngr.close()
	defer b.logMux.Close()
}
