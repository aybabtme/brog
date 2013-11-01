package main

import (
	"github.com/aybabtme/brog/brogger"
	"os"
)

const (
	// CreateOnly a brog structure, but don't run the brog.
	CreateOnly = "create"
)

func main() {

	brog, err := brogger.PrepareBrog()
	if err != nil {
		panic(err)
	}
	defer brog.Close()

	for _, arg := range os.Args[1:] {
		switch arg {
		case CreateOnly:
			brog.Ok("Only creating brog structure. Bye!")
			err := brogger.CopyBrogBinaries(brog.Config)
			if err != nil {
				brog.Err("Could not write sample brog post, %v", err)
			}
			return
		default:
			brog.Warn("Unknown command: %s, ignoring", arg)
		}
	}

	err = brog.ListenAndServe()
	brog.Err("Whoops! %v", err)
}
