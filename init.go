package main

import (
	"flag"
	"os"
	"strconv"
)

const (
	// DefaultPort on which the app listens.
	DefaultPort = 3000
	// PortEnvVar is the string used to get the port number from the environment
	PortEnvVar  = "BROG_PORT"
	PortFlagVar = "port"
)

var (
	port     = DefaultPort
	portFlag *int
)

func init() {
	portFlag = flag.Int(PortFlagVar, DefaultPort, "port number to listen to")
	flag.Parse()

	initPortVar()
}

func initPortVar() {

	isValid := func(p int) bool { return p > 0 && p < 1<<16 }

	// Port flag has precedence over environment variables
	if portFlag != nil && *portFlag != DefaultPort && isValid(*portFlag) {
		port = *portFlag
		sOut.Printf("Using '%s' flag", PortFlagVar)
		return
	}

	// Env var has precedence over Default value
	env := os.Getenv(PortEnvVar)
	portEnv, err := strconv.Atoi(env)
	if err == nil && isValid(portEnv) {
		port = portEnv
		sOut.Printf("Using '%s' variable", PortEnvVar)
	}
}
