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
	// Server starts brog at the current path
	Server = "server"
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
			doServer()
			return
		case Create:
			followingWords := strings.Join(commands[i+1:], "_")
			doCreate(followingWords)
			return
		default:
			printPreBrogError("Unknown command: %s.\n", arg)
		}
	}

}

func doInit() {
	fmt.Println(brush.DarkGray("A dark geometric shape is approaching..."))
	err := brogger.CopyBrogBinaries()
	if err != nil {
		printPreBrogError("Couldn't inject brog nanoprobes.\n")
		printPreBrogError("Message : %v.\n", err)
		return
	}

	brog, err := brogger.PrepareBrog()
	if err != nil {
		printPreBrogError("Couldn't prepare brog structure.\n")
		printPreBrogError("Message : %v.\n", err)
		return
	}
	brog.Ok("Initiliazing a brog. Resistance is futile.")

	defer closeOrPanic(brog)
	brog.Ok("Brog nanoprobes implanted.")
}

func doServer() {
	brog, err := brogger.PrepareBrog()
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

func doCreate(newPostFilename string) {
	brog, err := brogger.PrepareBrog()
	if err != nil {
		printPreBrogError("Couldn't create new post.\n")
		printPreBrogError("Message : %v.\n", err)
		printTryInitMessage()
		return
	}
	defer closeOrPanic(brog)

	err = brogger.CopyBlankToFilename(brog.Config, newPostFilename)
	if err != nil {
		brog.Err("Brog post creation failed, %v.", err)
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
