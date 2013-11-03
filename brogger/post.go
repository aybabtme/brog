package brogger

import (
	"encoding/json"
	"fmt"
	"github.com/russross/blackfriday"
	"io/ioutil"
	"net/url"
	"os"
	"time"
)

type post struct {
	filename string
	id       string

	Title     string    `json:"title"`
	Date      time.Time `json:"date"`
	Author    string    `json:"author"`
	Invisible bool      `json:"invisible"`
	Abstract  string    `json:"abstract"`
	Content   string    // Loaded from the Markdown part
}

func (p *post) GetID() string {
	return p.id
}

func (p *post) setID() {
	p.id = url.QueryEscape(stripExtension(p.filename))
}

func newPostFromFile(filename string) (*post, error) {

	post := post{filename: filename}

	postFile, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("opening file '%s', %v", filename, err)
	}

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

	post.setID()

	if err := postFile.Close(); err != nil {
		return &post, fmt.Errorf("closing post file, %v", err)
	}

	return &post, nil
}

type postList struct {
	posts []*post
}

func (p postList) Len() int {
	return len(p.posts)
}

func (p postList) Less(i, j int) bool {
	// Sort in most-recent order
	return p.posts[i].Date.After(p.posts[j].Date)
}

func (p postList) Swap(i, j int) {
	p.posts[i], p.posts[j] = p.posts[j], p.posts[i]
}
