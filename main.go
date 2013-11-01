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
			err := makeSample(brog.Config.PostPath, "sample.md")
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

func makeSample(dirpath, name string) error {
	return brogger.WriteSamplePost(dirpath + string(os.PathSeparator) + name)
}
