package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/aybabtme/brog/brogger"
	"github.com/aybabtme/log"
)

const (
	version = "devel"

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
	// Version shows the current version of brog
	Version = "version"

	usage = `usage: brog {init | server [prod] | create [new post name] | page [new page name] | version}

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

    brog version          Prints the current version of brog.
`
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
		case Version:
			fmt.Println(version)
			return
		case Help:
		default:
			log.KV("command", arg).Error("unknown command")
		}
	}
	fmt.Println(usage)
}

func doInit() {
	errs := brogger.CopyBrogBinaries()
	if len(errs) != 0 {
		log.KV("error.count", len(errs)).Error("multiple errors encountered")
		for _, err := range errs {
			log.Err(err).Error("initialization error")
		}
		return
	}

	brog, err := brogger.PrepareBrog(false)
	if err != nil {
		log.Err(err).Error("can't prepare brog")
		return
	}
	log.Info("initiliazing a brog, resistance is futile")

	defer closeOrPanic(brog)
}

func doServer(isProd bool) {

	brog, err := brogger.PrepareBrog(isProd)
	if err != nil {
		log.Err(err).Fatal("can't start brog server")
		return
	}
	defer closeOrPanic(brog)

	err = brog.ListenAndServe()
	log.Err(err).Fatal("failed to serve")
}

func doCreate(newPostFilename string, creationType string) {
	brog, err := brogger.PrepareBrog(false)
	if err != nil {
		return
	}
	defer closeOrPanic(brog)

	if creationType == "page" {
		err = brogger.CopyBlankToFilename(brog.Config, newPostFilename, brog.Config.PagePath)
	} else {
		err = brogger.CopyBlankToFilename(brog.Config, newPostFilename, brog.Config.PostPath)
	}
	if err != nil {
		log.Err(err).KV("creation.type", creationType).Error("creation of brog failed")
		return
	}
	log.KV("file.name", newPostFilename).Info("creation of brog post successful")
}

func closeOrPanic(brog *brogger.Brog) {
	if err := brog.Close(); err != nil {
		panic(err)
	}
}
