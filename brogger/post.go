package brogger

import (
	"encoding/json"
	"fmt"
	"github.com/russross/blackfriday"
	"io/ioutil"
	"os"
	"time"
)

type Post struct {
	filename  string
	Title     string    `json:"title"`
	Date      time.Time `json:"date"`
	Author    string    `json:"author"`
	Invisible bool      `json:"invisible"`
	Abstract  string    `json:"abstract"`
	Content   string
}

func NewPostFromFile(filename string) (*Post, error) {

	post := Post{filename: filename}

	postFile, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("opening file '%s', %v", filename, err)
	}
	defer postFile.Close()

	dec := json.NewDecoder(postFile)
	if err = dec.Decode(&post); err != nil {
		return nil, fmt.Errorf("reading json header of post '%s', %v", filename, err)
	}

	// Read content after that JSON header.  json.Decoder buffers some data, so
	// we need to ask for a buffered reader on it's current position.
	markdownContent, err := ioutil.ReadAll(dec.Buffered())
	if err != nil {
		return nil, fmt.Errorf("reading markdown content of post '%s', %v", post.Title, err)
	}

	htmlContent := blackfriday.MarkdownCommon(markdownContent)
	post.Content = string(htmlContent)

	return &post, nil
}

type PostList struct {
	posts []*Post
}

func (p PostList) Len() int {
	return len(p.posts)
}

func (p PostList) Less(i, j int) bool {
	// Sort in most-recent order
	return p.posts[i].Date.After(p.posts[j].Date)
}

func (p PostList) Swap(i, j int) {
	p.posts[i], p.posts[j] = p.posts[j], p.posts[i]
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
