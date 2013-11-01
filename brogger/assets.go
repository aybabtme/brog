package brogger

import (
	"io/ioutil"
	"os"
)

const (
	javascriptPath = "js"
	stylesheetPath = "css"
)

func CopyBrogBinaries(conf *Config) {
	os.MkdirAll(conf.PostPath, 0740)
	os.MkdirAll(conf.AssetPath, 0740)
	os.MkdirAll(conf.TemplatePath, 0740)
}

func WriteSamplePost(filename string) error {

	const samplePost = `{
    "title":"My First Post",
    "author":"Antoine Grondin",
    "date":"2013-10-30T23:59:59.000Z",
    "invisible": true,
    "abstract":"My first post using Brog"
}
# Hello!!
This is my first Brog post.  I really like broging and feeling like I'm finally one of those Broggers.  At last, I'm part of a community!

` + "```go" + `
func Hello() {
    fmt.Printf("Hello?")
}
` + "```" + `

Maybe!  __Who knows!!!__  Hopefully [this will be a link][1].

> Don't click it!!

_Shhhh_.

## Reasons why Antoine is great

* He has a nice beard.
* He has a nice way of mispeaking English.
* He prefers Star Trek to Star Wars.

[1]: en.wikipedia.org/wiki/Borg_(Star_Trek)
`

	return ioutil.WriteFile(filename, []byte(samplePost), 0640)
}
