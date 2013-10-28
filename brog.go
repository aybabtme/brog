package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
)

const (
	// DefaultPort on which the app listens.
	DefaultPort = 3000
	// PortEnvVar is the string used to get the port number from the environment
	PortEnvVar = "BORT_PORT"

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

var port = DefaultPort

func initPortVar() {

	isValid := func(p int) bool { return p > 0 && p < 1<<16 }

	portEnv, err := strconv.Atoi(os.Getenv(PortEnvVar))
	if err != nil && isValid(portEnv) {
		port = portEnv
	}

	portFlag := flag.Int("port", DefaultPort, "port number to listen to")
	if portFlag != nil && isValid(*portFlag) {
		port = *portFlag
	}
}

// HeartBeat answers 200 to any request.
func HeartBeat(rw http.ResponseWriter, req *http.Request) {
	rw.WriteHeader(http.StatusOK)
}

func Index(rw http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(rw, Placeholder)
}

func main() {

	initPortVar()

	http.HandleFunc("/heartbeat", HeartBeat)
	http.HandleFunc("/", Index)

	addr := fmt.Sprintf(":%d", port)
	log.Printf("Listening at %s", addr)
	http.ListenAndServe(addr, nil)
}
