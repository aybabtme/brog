package main

import (
	"fmt"
	"github.com/aybabtme/brog/brogger"
	"github.com/aybabtme/color/brush"
	"github.com/docopt/docopt.go"
	"os"
	"strings"
)

const (
	usage = `Brog

Usage:
  brog -i | --init
  brog (-s | --server) [--prod]
  brog (-c | --create) [--page] <new post name>...
  brog -h | --help
  brog -v | --version

Options:
  -i --init      Creates a new brog structure at the current working directory
  -s --server    Starts serving the brog structure at current location and watch
                 for changes in the template, page and post folders specified by
                 the config file. If --prod, use production port/socket
                 specified in the config file. Brog runs in development mode by
                 default
  -c --create    Creates a blank post in the file specified on the commandline
                 in the locations specified by the config file. Creates a blank
                 page instead of a post if --page.
  -v --version   Displays the current version of brog
  -h --help      Displays this help message
`
)

var (
	errPfx = fmt.Sprintf("%s%s%s ",
		brush.DarkGray("["),
		brush.Red("ERROR"),
		brush.DarkGray("]"))
)

func main() {
	arguments, _ := docopt.Parse(usage, nil, true, brogger.Version, false)
	if arguments["--init"].(bool) == true {
		doInit()
		return
	} else if arguments["--server"].(bool) == true {
		doServer(arguments["--prod"].(bool) == true)
		return
	} else if arguments["--create"].(bool) == true {
		newPostNameList := arguments["<new post name>"]
		followingWords := strings.Join(newPostNameList.([]string), "_")
		if arguments["--page"].(bool) == true {
			doCreate(followingWords, "page")
		} else {
			doCreate(followingWords, "post")
		}
		return
	} else if arguments["--version"].(bool) == true {
		fmt.Println(brogger.Version)
		return
	}
	fmt.Println(usage)
}

func doInit() {
	fmt.Println(brush.DarkGray("A dark geometric shape is approaching..."))
	errs := brogger.CopyBrogBinaries()
	if len(errs) != 0 {
		printPreBrogError("Couldn't inject brog nanoprobes.\n")
		for _, err := range errs {
			printPreBrogError("Message : %v.\n", err)
		}
		return
	}

	brog, err := brogger.PrepareBrog(false)
	if len(errs) != 0 {
		printPreBrogError("Couldn't prepare brog structure.\n")
		printPreBrogError("Message : %v.\n", err)
		return
	}
	brog.Ok("Initiliazing a brog. Resistance is futile.")

	defer closeOrPanic(brog)
	brog.Ok("Brog nanoprobes implanted.")
}

func doServer(isProd bool) {

	brog, err := brogger.PrepareBrog(isProd)
	if err != nil {
		printPreBrogError("Couldn't start brog server.\n")
		printPreBrogError("Message : %v.\n", err)
		printTryInitMessage()
		return
	}
	defer closeOrPanic(brog)

	err = brog.ListenAndServe()
	brog.Err("Whoops! %v.", err)

}

func doCreate(newPostFilename string, creationType string) {
	brog, err := brogger.PrepareBrog(false)
	if err != nil {
		printPreBrogError("Couldn't create new post.\n")
		printPreBrogError("Message : %v.\n", err)
		printTryInitMessage()
		return
	}
	defer closeOrPanic(brog)

	if creationType == "page" {
		err = brogger.CopyBlankToFilename(brog.Config, newPostFilename, brog.Config.PagePath)
	} else {
		err = brogger.CopyBlankToFilename(brog.Config, newPostFilename, brog.Config.PostPath)
	}
	if err != nil {
		brog.Err("Brog %s creation failed, %v.", creationType, err)
		brog.Err("Why do you resist?")
		return
	}
	brog.Ok("'%s' will become one with the Brog.", newPostFilename)
}

func printPreBrogError(format string, args ...interface{}) {
	errMsg := fmt.Sprintf("%s%s", errPfx, format)
	fmt.Fprintf(os.Stderr, errMsg, args...)
}

func printTryInitMessage() {
	fmt.Printf("Try initializing a brog here, run : brog --init.\n")
}

func closeOrPanic(brog *brogger.Brog) {
	if err := brog.Close(); err != nil {
		panic(err)
	}
}
