package main

import (
	"fmt"
	"github.com/aybabtme/brog/brogger"
	"net/http"
	"os"
)

const (

	// CreateOnly a brog structure, but don't run the brog.
	CreateOnly = "create"

	// Placeholder for now
	Placeholder = `<!DOCTYPE html>
<html>
<head><title>Brog Welcomes You</title></head>
<body>
    <h1>Hello From Brog</h1>
    <p>Brog loves you.  Have a nice day.</p>
</body>
</html>`
)

// HeartBeat answers 200 to any request.
func HeartBeat(rw http.ResponseWriter, req *http.Request) {
	rw.WriteHeader(http.StatusOK)
}

// Index serves placeholder for now
func Index(rw http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(rw, Placeholder)
}

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

	http.HandleFunc("/heartbeat", HeartBeat)
	http.HandleFunc("/", Index)

	err = brog.ListenAndServe()
	brog.Err("Whoops! %v", err)
}

func makeSample(dirpath, name string) error {
	return brogger.WriteSamplePost(dirpath + string(os.PathSeparator) + name)
}
