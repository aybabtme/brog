package brogger

import (
	"encoding/json"
	"fmt"
	"github.com/russross/blackfriday"
	"io/ioutil"
	"net/url"
	"os"
	"strings"
	"time"
)

type Post struct {
	filename string
	id       string

	Title     string    `json:"title"`
	Date      time.Time `json:"date"`
	Author    string    `json:"author"`
	Invisible bool      `json:"invisible"`
	Abstract  string    `json:"abstract"`
	Content   string    // Loaded from the Markdown part
}

func (p *Post) GetID() string {
	return p.id
}

func (p *Post) setID() {

	// TODO this is messing up the file watch
	date := fmt.Sprintf("%d_%s_%d_", p.Date.Day(), p.Date.Month().String(), p.Date.Year())

	cleanTitle := strings.TrimSpace(p.Title)
	lowerTitle := strings.ToLower(cleanTitle)
	snakeTitle := strings.Replace(lowerTitle, " ", "_", -1)

	rawID := strings.TrimSpace(date + snakeTitle)

	p.id = url.QueryEscape(rawID)
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

	post.setID()

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
