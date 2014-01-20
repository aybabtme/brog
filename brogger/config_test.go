package brogger

import (
	"io"
	"os"
	"testing"
)

var config *Config

func TestSelfValidate(t *testing.T) {
	config = newDefaultConfig()
	err := config.selfValidate()
	if err != nil {
		t.Error("Error validating default conig")
	}
	config.ProdPort = -1
	err = config.selfValidate()
	if err == nil {
		t.Error("-1 is not a valid production port number")
	}
	config.ProdPort = 65600
	err = config.selfValidate()
	if err == nil {
		t.Error("656000 is not a valid production port number")
	}
	config.ProdPort = DefaultProdPort
	config.DevelPort = -1
	err = config.selfValidate()
	if err == nil {
		t.Error("-1 is not a valid development port number")
	}
	config.DevelPort = 65600
	err = config.selfValidate()
	if err == nil {
		t.Error("65600 is not a valid development port number")
	}
	config.DevelPort = DefaultDevelPort
	config.MaxCPUs = -1
	err = config.selfValidate()
	if err == nil {
		t.Error("-1 is not a valid number of threads")
	}
	config.MaxCPUs = DefaultMaxCPUs
	config.PostFileExt = ""
	err = config.selfValidate()
	if err == nil {
		t.Error("Post file extension cannot be empty")
	}
}

func TestJsonConfigStruct(t *testing.T) {
	os.Chdir("base")
	defer os.Chdir("..")
	config, _ = loadConfig()
	config.persistToFile("test_config.json")
	defer os.Remove("test_config.json")
	origfile, _ := os.Open("brog_config.json")
	testfile, _ := os.Open("test_config.json")
	origbuf := make([]byte, 512)
	testbuf := make([]byte, 512)
	for {
		i, origerr := origfile.Read(origbuf)
		j, testerr := testfile.Read(testbuf)
		if origerr == io.EOF || testerr == io.EOF {
			if !(origerr == io.EOF && testerr == io.EOF) {
				t.Error("One config file ended before the other")
			}
			break
		}
		if i != j {
			t.Error("Different amounts read into each config file buffer:", i, j)
			break
		}
		for k := 0; k < i; k++ {
			if origbuf[k] != testbuf[k] {
				t.Error("Config files are not byte-by-byte equal")
			}
		}
	}
}
