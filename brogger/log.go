package brogger

import (
	"fmt"
	"github.com/aybabtme/color/brush"
	"log"
	"os"
)

type logMux struct {
	logFile *os.File

	okFile    *log.Logger
	okConsole *log.Logger

	warnFile    *log.Logger
	warnConsole *log.Logger

	errorFile    *log.Logger
	errorConsole *log.Logger
}

func makeLogMux(filename string) (*logMux, error) {
	file, err := os.OpenFile(filename, os.O_APPEND|os.O_WRONLY, 0640)
	if os.IsNotExist(err) {
		file, err = os.Create(filename)
	}
	if err != nil {
		return nil, fmt.Errorf("opening log file %s, %v", filename, err)
	}

	return &logMux{
		logFile:      file,
		okFile:       log.New(file, "[OK]   ", log.LstdFlags),
		warnFile:     log.New(file, "[WARN] ", log.LstdFlags),
		errorFile:    log.New(file, "[ERR]  ", log.LstdFlags),
		okConsole:    log.New(os.Stdout, brush.Green("[OK]   ").String(), log.LstdFlags),
		warnConsole:  log.New(os.Stdout, brush.Yellow("[WARN] ").String(), log.LstdFlags),
		errorConsole: log.New(os.Stderr, brush.Red("[ERR]  ").String(), log.LstdFlags),
	}, nil
}

func (l *logMux) Ok(format string, args ...interface{}) {
	l.okConsole.Printf(format, args...)
	l.okFile.Printf(format, args...)
}

func (l *logMux) Warn(format string, args ...interface{}) {
	l.warnConsole.Printf(format, args...)
	l.warnFile.Printf(format, args...)
}
func (l *logMux) Err(format string, args ...interface{}) {
	l.errorConsole.Printf(format, args...)
	l.errorFile.Printf(format, args...)
}

func (l *logMux) Close() error {
	if err := l.logFile.Close(); err != nil {
		return fmt.Errorf("closing log file, %v", err)
	}
	return nil
}
