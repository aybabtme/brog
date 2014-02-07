package main

import (
	"fmt"
	"github.com/aybabtme/brog/brogger"
	"github.com/aybabtme/color/brush"
	"os"
	"strings"
)

const (
	// Init a brog structure, but don't run the brog.
	Init = "init"
	// Create a new blank post in the post folder
	Create = "create"
	// Create a new blank page in the page folder
	Page = "page"
	// Server starts brog at the current path
	Server = "server"
	// Help shows the usage string
	Help = "help"

	usage = `usage: brog {init | server [prod] | create [new post name] | page [new page name]}

'brog' is a tool to initialize brog structures, serve the content
of brog structures and create new posts in a brog structure.

The following are brog's valid commands with the arguments they take :

    brog init             Takes no argument, creates a new brog struc-
                          ture at the current working directory.

    brog server [prod]    Starts serving the brog structure at the
                          current location and watch for changes in the
                          template and post folders specified by the
                          config file.  If [prod], use the production
                          port number specified in the config file. By
                          default, brog runs in development mode.

    brog create [name]    Creates a blank post in file [name], in the
                          location specified by the config file.

    brog page [name]      Creates a blank page in file [name], in the
                          location specified by the config file.

    brog help             Shows this message.
`
)

var (
	errPfx = fmt.Sprintf("%s%s%s ",
		brush.DarkGray("["),
		brush.Red("ERROR"),
		brush.DarkGray("]"))
)

func main() {
	commands := os.Args[1:]
	for i, arg := range commands {
		switch arg {
		case Init:
			doInit()
			return
		case Server:
			if len(commands) > i+1 {
				doServer(commands[i+1] == "prod")
			} else {
				doServer(false)
			}
			return
		case Create:
			followingWords := strings.Join(commands[i+1:], "_")
			doCreate(followingWords, "post")
			return
		case Page:
			followingWords := strings.Join(commands[i+1:], "_")
			doCreate(followingWords, "page")
			return
		case Help:
		default:
			printPreBrogError("Unknown command: %s.\n", arg)
		}
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
	fmt.Printf("Try initializing a brog here, run : brog %s.\n", Init)
}

func closeOrPanic(brog *brogger.Brog) {
	if err := brog.Close(); err != nil {
		panic(err)
	}
}
