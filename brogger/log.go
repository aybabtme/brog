package brogger

import (
	"fmt"
	"github.com/aybabtme/color/brush"
	"log"
	"os"
)

// Verbosity levels accepted by brog
const (
	DebugVerbosity = "debug"
	WatchVerbosity = "watch"
	InfoVerbosity  = "info"
	WarnVerbosity  = "warn"
	ErrorVerbosity = "error"
)

const (
	errorLevel int = iota
	warnLevel
	infoLevel
	watchLevel
	debugLevel
)

func getVerbosityLevel(verbStr string) (int, error) {
	switch verbStr {
	case DebugVerbosity:
		return debugLevel, nil
	case WatchVerbosity:
		return watchLevel, nil
	case InfoVerbosity:
		return infoLevel, nil
	case WarnVerbosity:
		return warnLevel, nil
	case ErrorVerbosity:
		return errorLevel, nil
	default:
		return 0, fmt.Errorf("'%s' is not a verbosity level", verbStr)
	}
}

type logMux struct {
	logFile *os.File

	debugFile    *log.Logger
	debugConsole *log.Logger

	watchFile    *log.Logger
	watchConsole *log.Logger

	infoFile    *log.Logger
	infoConsole *log.Logger

	warnFile    *log.Logger
	warnConsole *log.Logger

	errorFile    *log.Logger
	errorConsole *log.Logger

	fileVerbose    int
	consoleVerbose int
}

func makeLogMux(conf *Config) (*logMux, error) {

	fileVerb, err := getVerbosityLevel(conf.LogFileVerbosity)
	if err != nil {
		return nil, fmt.Errorf("log file verbosity, %v", err)
	}
	consoleVerb, err := getVerbosityLevel(conf.ConsoleVerbosity)
	if err != nil {
		return nil, fmt.Errorf("console verbosity, %v", err)
	}

	filename := conf.LogFilename

	file, err := os.OpenFile(filename, os.O_APPEND|os.O_WRONLY, 0640)
	if os.IsNotExist(err) {
		file, err = os.Create(filename)
	}
	if err != nil {
		return nil, fmt.Errorf("opening log file %s, %v", filename, err)
	}

	debugPfx := fmt.Sprintf("%s%s%s ",
		brush.DarkGray("["),
		brush.Blue("DEBUG"),
		brush.DarkGray("]"))

	watchPfx := fmt.Sprintf("%s%s%s ",
		brush.DarkGray("["),
		brush.DarkCyan("WATCH"),
		brush.DarkGray("]"))

	infoPfx := fmt.Sprintf(" %s%s%s ",
		brush.DarkGray("["),
		brush.Green("INFO"),
		brush.DarkGray("]"))

	warnPfx := fmt.Sprintf(" %s%s%s ",
		brush.DarkGray("["),
		brush.Yellow("WARN"),
		brush.DarkGray("]"))

	errPfx := fmt.Sprintf("%s%s%s ",
		brush.DarkGray("["),
		brush.Red("ERROR"),
		brush.DarkGray("]"))

	return &logMux{
		logFile:        file,
		debugFile:      log.New(file, "[DEBUG] ", log.LstdFlags),
		watchFile:      log.New(file, "[WATCH] ", log.LstdFlags),
		infoFile:       log.New(file, " [INFO] ", log.LstdFlags),
		warnFile:       log.New(file, " [WARN] ", log.LstdFlags),
		errorFile:      log.New(file, "[ERROR] ", log.LstdFlags),
		debugConsole:   log.New(os.Stdout, debugPfx, log.LstdFlags),
		watchConsole:   log.New(os.Stdout, watchPfx, log.LstdFlags),
		infoConsole:    log.New(os.Stdout, infoPfx, log.LstdFlags),
		warnConsole:    log.New(os.Stdout, warnPfx, log.LstdFlags),
		errorConsole:   log.New(os.Stderr, errPfx, log.LstdFlags),
		fileVerbose:    fileVerb,
		consoleVerbose: consoleVerb,
	}, nil
}

func (l *logMux) Debug(format string, args ...interface{}) {
	if l.consoleVerbose >= debugLevel {
		l.debugConsole.Printf(format, args...)
	}
	if l.fileVerbose >= debugLevel {
		l.debugFile.Printf(format, args...)
	}
}

func (l *logMux) Watch(format string, args ...interface{}) {
	if l.consoleVerbose >= watchLevel {
		l.watchConsole.Printf(format, args...)
	}
	if l.fileVerbose >= watchLevel {
		l.watchFile.Printf(format, args...)
	}
}

func (l *logMux) Ok(format string, args ...interface{}) {
	if l.consoleVerbose >= infoLevel {
		l.infoConsole.Printf(format, args...)
	}
	if l.fileVerbose >= infoLevel {
		l.infoFile.Printf(format, args...)
	}
}

func (l *logMux) Warn(format string, args ...interface{}) {
	if l.consoleVerbose >= warnLevel {
		l.warnConsole.Printf(format, args...)
	}
	if l.fileVerbose >= warnLevel {
		l.warnFile.Printf(format, args...)
	}
}
func (l *logMux) Err(format string, args ...interface{}) {
	if l.consoleVerbose >= errorLevel {
		l.errorConsole.Printf(format, args...)
	}
	if l.fileVerbose >= errorLevel {
		l.errorFile.Printf(format, args...)
	}
}

func (l *logMux) Close() error {
	if err := l.logFile.Close(); err != nil {
		return fmt.Errorf("closing log file, %v", err)
	}
	return nil
}
