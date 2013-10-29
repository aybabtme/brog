package brogger

import (
	"fmt"
	"net/http"
	"os"
	"runtime"
)

const ()

type Brog struct {
	*logMux
	Config *Config
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

	brog := Brog{
		logMux: logMux,
		Config: config,
	}

	runtime.GOMAXPROCS(config.MaxCPUs)
	brog.watchTemplates(config.TemplatePath)
	brog.watchPosts(config.PostPath)

	return &brog, nil
}

func (b *Brog) ListenAndServe() error {
	addr := fmt.Sprintf("%s:%d", b.Config.Hostname, b.Config.PortNumber)
	b.Ok("Borg open for business on %s", addr)
	return http.ListenAndServe(addr, nil)
}

func (b *Brog) watchTemplates(templPath string) {
	os.MkdirAll(templPath, 0740)
	b.Warn("Not implemented yet! Watch path at %s", templPath)
}

func (b *Brog) watchPosts(postPath string) {
	os.MkdirAll(postPath, 0740)
	b.Warn("Not implemented yet! Watch path at %s", postPath)
}
