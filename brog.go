package main

import (
	"fmt"
	"github.com/aybabtme/color/brush"
	"log"
	"net/http"
	"os"
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

var (
	sOut = log.New(os.Stdout, brush.Green("[OK]  ").String(), log.LstdFlags)
	sErr = log.New(os.Stderr, brush.Red("[ERR] ").String(), log.LstdFlags)
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

	addr := fmt.Sprintf("localhost:%d", port)

	sOut.Printf("Borg open for business on %s", addr)
	err := http.ListenAndServe(addr, nil)
	sErr.Printf("Whoops! %v", err)
}
