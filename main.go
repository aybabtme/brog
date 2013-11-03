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
			doCreate(followingWords + ".md")
			return
		default:
			fmt.Printf("Unknown command: %s\n", arg)
		}
	}

}

func doInit() {
	fmt.Println(brush.DarkGray("A dark geometric shape is approaching..."))
	err := brogger.CopyBrogBinaries()
	if err != nil {
		panic(err)
	}

	brog, err := brogger.PrepareBrog()
	if err != nil {
		panic(err)
	} else {
		brog.Ok("Initiliazing a brog. Resistance is futile.")
	}
	defer closeOrPanic(brog)
	brog.Ok("Brog nanoprobes implanted.")
}

func doServer() {
	brog, err := brogger.PrepareBrog()
	if err != nil {
		fmt.Printf("%s %v", errPfx, err)
		return
	}
	defer closeOrPanic(brog)

	err = brog.ListenAndServe()
	brog.Err("Whoops! %v", err)

}

func doCreate(newPostFilename string) {
	brog, err := brogger.PrepareBrog()
	if err != nil {
		fmt.Printf("%s %v", errPfx, err)
		return
	}
	defer closeOrPanic(brog)

	brog.Ok("Brog post '%s' will be assimilated.", newPostFilename)
	err = brogger.CopyBlankToFilename(brog.Config, newPostFilename)
	if err != nil {
		brog.Err("Brog post creation failed, %v", err)
		brog.Err("Why do you resist?")
		return
	}
	brog.Ok("You will become one with the Borg.")
}

func closeOrPanic(brog *brogger.Brog) {
	if err := brog.Close(); err != nil {
		panic(err)
	}
}
