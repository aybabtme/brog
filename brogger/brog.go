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

func (b *Brog) ListenAndServe() error {
	addr := fmt.Sprintf("%s:%d", b.Config.Hostname, b.Config.PortNumber)
	b.Ok("Borg open for business on %s", addr)
	return http.ListenAndServe(addr, nil)
}
