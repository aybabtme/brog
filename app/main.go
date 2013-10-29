package main

import (
	"fmt"
	"github.com/aybabtme/brog"
	"net/http"
)

const (
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

	http.HandleFunc("/heartbeat", HeartBeat)
	http.HandleFunc("/", Index)

	borg, err := brog.PrepareBrog()
	if err != nil {
		panic(err)
	}

	err = borg.ListenAndServe()
	borg.Err("Whoops! %v", err)
}
