package main

import (
	"fmt"
	"github.com/aybabtme/brog/brogger"
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
			fmt.Printf("Unknown command: %s\n", arg)
		}
	}

}

func doInit() {

	brog, err := brogger.PrepareBrog()
	if err != nil {
		panic(err)
	}
	defer brog.Close()

	brog.Ok("Initiliazing a brog. Resistance is futile.")
	err = brogger.CopyBrogBinaries(brog.Config)
	if err != nil {
		brog.Err("Assimilation failed, %v", err)
	} else {
		brog.Ok("Brog nanoprobes implanted.")
	}
}

func doServer() {
	brog, err := brogger.PrepareBrog()
	if err != nil {
		panic(err)
	}
	defer brog.Close()

	err = brog.ListenAndServe()
	brog.Err("Whoops! %v", err)
}

func doCreate(newPostFilename string) {
	brog, err := brogger.PrepareBrog()
	if err != nil {
		panic(err)
	}
	defer brog.Close()

	brogger.CopyBlankToFilename(brog.Config, newPostFilename)
}
