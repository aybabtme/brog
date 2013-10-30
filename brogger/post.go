package brogger

import (
	"encoding/json"
	"fmt"
	"github.com/russross/blackfriday"
	"io/ioutil"
	"os"
	"path"
	"time"
)

type PostManager struct {
	brog  *Brog
	path  string
	posts map[string]*Post
}

func StartPostManager(brog *Brog, filepath string) (*PostManager, error) {

	brog.Warn("Not implemented yet! Watch path at %s", filepath)

	os.MkdirAll(filepath, 0740)
	fileInfos, err := ioutil.ReadDir(filepath)
	if err != nil {
		return nil, fmt.Errorf("listing directory '%s', %v", filepath, err)
	}

	postMngr := &PostManager{
		brog:  brog,
		path:  filepath,
		posts: make(map[string]*Post),
	}

	for _, fileInfo := range fileInfos {
		if !fileInfo.IsDir() {
			err := postMngr.loadFromFile(filepath, fileInfo.Name())
			if err != nil {
				brog.Warn("Loading post from file '%s' failed, %v", fileInfo.Name(), err)
			}
		}
	}

	return postMngr, nil
}

func (t *PostManager) loadFromFile(dirname, filename string) error {
	t.brog.Warn("Not implemented yet! Load post from file %s", filename)

	post, err := NewPostFromFile(dirname, filename)
	if err != nil {
		return fmt.Errorf("loading post from file '%s', %v", filename, err)
	}

	t.posts[filename] = post
	t.brog.Ok("Loaded post '%s'", post.Title)

	return nil
}

type Post struct {
	filename string
	Title    string    `json:"title"`
	Date     time.Time `json:"date"`
	Author   string    `json:"author"`
	Abstract string    `json:"abstract"`
	Content  string
}

func NewPostFromFile(dirname, filename string) (*Post, error) {

	fullPath := fmt.Sprintf("%s%c%s", path.Clean(dirname), os.PathSeparator, filename)

	post := Post{filename: fullPath}

	postFile, err := os.Open(fullPath)
	if err != nil {
		return nil, fmt.Errorf("opening file '%s', %v", fullPath, err)
	}
	defer postFile.Close()

	dec := json.NewDecoder(postFile)
	if err = dec.Decode(&post); err != nil {
		return nil, fmt.Errorf("reading json header of post '%s', %v", fullPath, err)
	}

	// Read content after that JSON header.  json.Decoder buffers some data, so
	// we need to ask for a buffered reader on it's current position.
	markdownContent, err := ioutil.ReadAll(dec.Buffered())
	if err != nil {
		return nil, fmt.Errorf("reading markdown content of post '%s', %v", post.Title, err)
	}

	htmlContent := blackfriday.MarkdownCommon(markdownContent)
	post.Content = string(htmlContent)

	fmt.Printf("\n\n%s\n\n", post.Content)

	return &post, nil
}

func WriteSamplePost(filename string) error {

	const samplePost = `{
    "title":"My First Post",
    "author":"Antoine Grondin",
    "date":"2013-10-30T23:59:59.000Z",
    "abstract":"My first post using Brog"
}
# Hello!!
This is my first Brog post.  I really like broging and feeling like I'm finally one of those Broggers.  At last, I'm part of a community!

` + "```" + `
Fenced code block?
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
